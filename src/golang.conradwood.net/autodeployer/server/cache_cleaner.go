package main

import (
	"flag"
	"fmt"
	"golang.conradwood.net/autodeployer/deployments"
	"golang.conradwood.net/autodeployer/downloader"
	"golang.conradwood.net/go-easyops/utils"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	enable_cleaner    = flag.Bool("enable_cache_cleaner", true, "periodically clean cache if it grows too big")
	max_cache_size_gb = flag.Int("max_cache_size", 10, "maximum cache size to maintain in `gigabytes`. LRU applies")
)

func CacheStart() {
	go refresh_running_cache()
	go cache_clean_loop()
}

// purpose: tell on-disk-cache that a url is currently running (so it won't get evicted)
func refresh_running_cache() {
	for {
		time.Sleep(30 * time.Second)
		for _, d := range deployments.ActiveDeployments() {
			downloader.URLActive(d.URL())

		}
	}
}

func cache_clean_loop() {
	d := time.Duration(5) * time.Second
	for {
		time.Sleep(d)
		d = time.Duration(5) * time.Minute
		if !*enable_cleaner || *cache_v2 {
			continue
		}
		if time.Since(lastByteDownloaded) < time.Duration(5)*time.Minute {
			// last download was quite recent, don't clean just yet
			continue
		}
		err := cleanDownloadCache()
		if err != nil {
			fmt.Printf("Failed to clean cache: %s\n", err)
		}
	}
}

func cleanDownloadCache() error {
	downloadLock.Lock()
	defer downloadLock.Unlock()
	if !*enable_cleaner {
		return nil
	}
	if *cache_v2 {
		panic("v1 cache should be disabled!")
	}

	if time.Since(lastByteDownloaded) < time.Duration(5)*time.Minute {
		// last download was quite recent, don't clean just yet
		return nil
	}
	ce, err := ReadCacheFile()
	if err != nil {
		return err
	}

	dfiles, err := readDir()
	if err != nil {
		return err
	}

	for _, d := range dfiles {
		if ce.FindFile(d) != nil {
			continue
		}
		c := &CacheEntry{
			tobedeleted: true,
			created:     time.Unix(1000000, 0),
			url:         "deleteme",
			filename:    d,
			size:        -1,
		}
		ce.entries = append(ce.entries, c)
	}

	sort.Slice(ce.entries, func(i, j int) bool { //we want earliest at highest index (ascending)
		return ce.entries[i].created.Before(ce.entries[j].created)
	})
	MAX_CACHE_SIZE := int64(*max_cache_size_gb) * 1024 * 1024 * 1024
	for ce.GetSize() > MAX_CACHE_SIZE {
		for _, e := range ce.entries {
			if e.tobedeleted {
				continue
			}
			e.tobedeleted = true
			break
		}
	}

	mustrewrite := false
	for _, e := range ce.entries {
		if !e.tobedeleted {
			continue
		}
		mustrewrite = true
		fmt.Printf("delete cache: Created: %s File: %s\n", utils.TimeString(e.created), e.filename)
		err := os.Remove(e.filename)
		if err != nil {
			fmt.Printf("Failed to delete: %s\n", err)
		}
	}

	if mustrewrite {
		err := ce.WriteFile()
		if err != nil {
			return err
		}
	}

	return nil
}

type CacheFile struct {
	entries []*CacheEntry
}

func ReadCacheFile() (*CacheFile, error) {
	if *cache_v2 {
		panic("v1 cache should be disabled!")
	}
	res := &CacheFile{}
	if !utils.FileExists(idxname) {
		return res, nil
	}
	s := ""
	ci, err := utils.ReadFile(idxname)
	if err != nil {
		return nil, err
	}
	s = string(ci)
	for i, l := range strings.Split(s, "\n") {
		if l == "" {
			continue
		}
		fields := strings.SplitN(l, " ", 3)
		if len(fields) != 3 {
			fmt.Printf("Invalid line %d in index file: \"%s\"\n", i, l)
			continue
		}
		ts, err := strconv.Atoi(fields[0])
		if err != nil {
			fmt.Printf("Invalid line %d in index file: \"%s\" (%s)\n", i, l, err)
			continue
		}

		filename := fields[1]
		curl := fields[2]
		ce := &CacheEntry{
			created:  time.Unix(int64(ts), 0),
			url:      curl,
			filename: filename,
			size:     -1,
		}
		res.entries = append(res.entries, ce)

	}
	return res, nil
}
func (c *CacheFile) FindFile(fname string) *CacheEntry {
	if *cache_v2 {
		panic("v1 cache should be disabled!")
	}

	for _, ce := range c.entries {
		if ce.filename == fname {
			return ce
		}
	}
	return nil
}

func (c *CacheFile) WriteFile() error {
	if *cache_v2 {
		panic("v1 cache should be disabled!")
	}
	s := ""
	for _, e := range c.entries {
		if e.tobedeleted {
			continue
		}
		s = s + fmt.Sprintf("%d %s %s\n", e.created.Unix(), e.filename, e.url)
	}
	err := utils.WriteFile(idxname, []byte(s))
	if err != nil {
		return err
	}
	return nil
}

func (c *CacheFile) GetSize() int64 {
	res := int64(0)
	for _, e := range c.entries {
		if e.tobedeleted {
			continue
		}
		i := e.getSize()
		if i == -1 {
			continue
		}
		res = res + i
	}
	return res
}

type CacheEntry struct {
	tobedeleted bool
	created     time.Time
	url         string
	filename    string
	size        int64 // -1 : not read yet
}

func (ce *CacheEntry) getSize() int64 {
	if ce.size != -1 {
		return ce.size
	}

	fi, err := os.Stat(ce.filename)
	if err != nil {
		if os.IsNotExist(err) {
			ce.tobedeleted = true
		}
		fmt.Printf("Failed to stat %s: %s\n", ce.filename, err)
		return -1
	}

	ce.size = fi.Size()
	return ce.size
}

func readDir() ([]string, error) {
	files, err := ioutil.ReadDir("/srv/autodeployer/download_cache")
	if err != nil {
		return nil, err
	}
	var res []string
	for _, f := range files {
		res = append(res, "/srv/autodeployer/download_cache/"+f.Name())
	}
	return res, nil
}
