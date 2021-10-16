package downloader

import (
	"flag"
	"fmt"
	"golang.conradwood.net/go-easyops/prometheus"
	"golang.conradwood.net/go-easyops/utils"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	max_cache_size = flag.Uint64("max_v2_cache_size_mb", 2000, "max size in megabytes for the v2 cache")
	flush_lock     sync.Mutex
	flush_chan     = make(chan *DiskCache)
	started        = time.Now()

	readersGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "autodeployer_cache_readers",
			Help: "V=1 UNIT=ops DESC=number of readers reading cache",
		})
	partialGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "autodeployer_cache_ramsize",
			Help: "V=1 UNIT=ops DESC=number of bytes partially cached in ram",
		})
)

// maintain the cache, regularly flush the index file to disk,
// clean out "old" entries etc
func (dc *DiskCache) cache_maintainer() {
	for {
		time.Sleep(time.Duration(2) * time.Second)
		err := dc.maintain_cache()
		if err != nil {
			fmt.Printf("error maintaining cache: %s\n", err)
		}
	}
}
func (dc *DiskCache) request_flushing() {
	flush_lock.Lock()
	dc.needs_flushing++
	flush_lock.Unlock()
}
func (dc *DiskCache) maintain_cache() error {
	dc.lock.Lock()
	err := dc.readonce()
	if err != nil {
		fmt.Printf("Failed to read index: %s\n", err)
	}
	dc.lock.Unlock()

	dc.set_stats()
	if !isEnabled() {
		err := dc.delete_quickly()
		if err != nil {
			return err
		}
	}

	err = dc.remove_inram()
	if err != nil {
		return err
	}

	err = dc.maintain_flush()
	if err != nil {
		return err
	}
	err = dc.delete_stale()
	if err != nil {
		return err
	}

	err = dc.delete_toobig()
	if err != nil {
		return err
	}

	err = dc.remove_untracked_files()
	if err != nil {
		return err
	}
	return nil
}

// remove files on disk that are not in the index
func (dc *DiskCache) remove_untracked_files() error {
	files, err := ioutil.ReadDir(CACHEDIR)
	if err != nil {
		return err
	}
	flush_lock.Lock()
	defer flush_lock.Unlock()

	for _, f := range files {
		fname := CACHEDIR + "/" + f.Name()
		//fmt.Printf("File: \"%s\"\n", fname)
		found := false
		for _, ce := range dc.entries {
			if ce.filename == fname {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Removing untracked file %s\n", fname)
			err := os.Remove(fname)
			if err != nil {
				fmt.Printf("failed to remove untracked file: %s\n", err)
			}
		}
	}
	return nil
}
func (dc *DiskCache) set_stats() error {
	flush_lock.Lock()
	defer flush_lock.Unlock()
	ramsize := uint64(0)
	readers := 0
	for _, ce := range dc.entries {
		ramsize = ramsize + uint64(len(ce.partial_data))
		readers = readers + len(ce.readers)
	}
	partialGauge.Set(float64(ramsize))
	readersGauge.Set(float64(readers))
	return nil
}

// remove buffers in byte, once file is on disk
func (dc *DiskCache) remove_inram() error {
	flush_lock.Lock()
	defer flush_lock.Unlock()
	for _, ce := range dc.entries {
		if !ce.completely_downloaded {
			continue
		}
		if len(ce.readers) == 0 && len(ce.partial_data) != 0 {
			ce.partial_data = nil
		}
	}
	return nil
}

func (dc *DiskCache) maintain_flush() error {
	flush_lock.Lock()
	defer flush_lock.Unlock()
	if dc.needs_flushing == 0 {
		return nil
	}
	err := dc.flush()
	if err != nil {
		return err
	}
	dc.needs_flushing = 0
	return nil
}

type toobig_check struct {
	entry *diskCacheEntry
	size  uint64
}

// reduce size of cache by removing least recently used
func (dc *DiskCache) delete_toobig() error {
	flush_lock.Lock()
	defer flush_lock.Unlock()
	// read size of cache
	var tc []*toobig_check
	for _, ce := range dc.entries {
		if !ce.completely_downloaded || ce.tobedeleted {
			continue
		}
		tb := &toobig_check{entry: ce}
		fi, err := os.Stat(ce.filename)
		if err != nil {
			return err
		}
		tb.size = uint64(fi.Size())
		tc = append(tc, tb)
	}
	size := uint64(0)
	for _, tb := range tc {
		size = size + tb.size
	}
	max_size := *max_cache_size * 1024 * 1024
	//	fmt.Printf("Cache Size: %s vs max %s\n", utils.PrettyNumber(size), utils.PrettyNumber(max_size))
	if size < max_size {
		return nil
	}
	fmt.Printf("Cache too big. Size: %s vs max %s\n", utils.PrettyNumber(size), utils.PrettyNumber(max_size))
	sort.Slice(tc, func(i, j int) bool {
		return tc[i].entry.lastRead.Before(tc[j].entry.lastRead)
	})
	for _, tb := range tc {
		fmt.Printf("Checking %s (%v)\n", tb.entry.url, tb.entry.lastRead)
	}
	new_size := size
	for _, tb := range tc {
		// don't delete stuff that's been read 30 seconds or less ago
		if time.Since(tb.entry.lastRead) < time.Duration(30)*time.Second {
			continue
		}
		fmt.Printf("Evicting %s\n", tb.entry.url)
		tb.entry.tobedeleted = true
		new_size = new_size - tb.size
		if new_size <= max_size {
			break
		}
	}
	return nil
}

// if cache is not enabled delete all that have not been used _very_ recently
func (dc *DiskCache) delete_quickly() error {
	// start deleting stuff 15 minutes after autodeployer booted.
	// this is to prevent the cache being deleted 2 minutes after boot and
	// instead give deploymonkey a chance to detect autodeployer, send it startup messages,
	// which then might be served from cache
	if time.Since(started) < time.Duration(15)*time.Minute {
		return nil
	}
	flush_lock.Lock()
	defer flush_lock.Unlock()
	// now check for failed downloads:
	for _, ce := range dc.entries {
		del := false
		if (ce.completely_downloaded) && time.Since(ce.lastRead) > time.Duration(120)*time.Second {
			del = true
		}

		if del {
			ce.abort_download = true
			ce.tobedeleted = true
			dc.needs_flushing++
			fmt.Printf("Marked as tobedeleted: %s\n", ce.filename)
		}
	}
	return nil
}

// delete stale ones, that is incomplete ones that were not updated for a while, or those which do not exist on disk
func (dc *DiskCache) delete_stale() error {
	flush_lock.Lock()
	defer flush_lock.Unlock()
	// now check for failed downloads:
	for _, ce := range dc.entries {
		del := false

		if time.Since(ce.lastModified) > time.Duration(30)*time.Second && !utils.FileExists(ce.filename) {
			del = true
		}

		if (!ce.completely_downloaded) && time.Since(ce.lastModified) > time.Duration(30)*time.Second {
			del = true
		}

		if del {
			ce.abort_download = true
			ce.tobedeleted = true
			dc.needs_flushing++
			fmt.Printf("Marked as tobedeleted: %s\n", ce.filename)
		}
	}

	for _, ce := range dc.entries {
		if !ce.tobedeleted {
			continue
		}
		fmt.Printf("Removing file %s\n", ce.filename)
		err := os.Remove(ce.filename)
		if err != nil {
			fmt.Printf("Failed to delete file: %s\n", err)
		}
	}

	dc.lock.Lock()
	var nentries []*diskCacheEntry
	o := len(dc.entries)
	for _, ce := range dc.entries {
		if ce.tobedeleted {
			continue
		}
		nentries = append(nentries, ce)
	}
	dc.entries = nentries
	if o != len(dc.entries) {
		dc.needs_flushing++
	}
	dc.lock.Unlock()
	return nil
}

func (dc *DiskCache) flush() error {
	flush_chan <- dc
	return nil
}

func flush_thread() {
	for {
		dc := <-flush_chan
		fmt.Printf("Cache flush...\n")
		var buf []string
		for _, ce := range dc.entries {
			buf = append(buf, ce.SerialiseToString())
		}
		fmt.Printf("Flushing %d entries\n", len(dc.entries))
		bs := strings.Join(buf, "\n") + "\n"
		bs = `#LastMod    filename                          URL                                               lastread  tobedeleted complDownloaded
` + bs
		//fmt.Printf("Todisk: %s\n", bs)
		err := utils.WriteFile(INDEXFILE, []byte(bs))
		if err != nil {
			fmt.Printf("Error writing file: %s\n", err)
		}
	}
}
