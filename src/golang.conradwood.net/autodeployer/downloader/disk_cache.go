package downloader

import (
	"crypto/md5"
	"fmt"
	"golang.conradwood.net/go-easyops/utils"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CACHEDIR  = "/srv/autodeployer/download_cache"
	INDEXFILE = "/srv/autodeployer/cache.index"
)

type DiskCache struct {
	lock           *sync.Mutex
	entries        []*diskCacheEntry
	needs_flushing int
}

func NewDiskCache() *DiskCache {
	res := &DiskCache{lock: &sync.Mutex{}}
	go res.cache_maintainer()
	return res
}

type diskCacheEntry struct {
	cache                 *DiskCache
	lastModified          time.Time
	lastRead              time.Time
	filename              string
	writefd               *os.File
	url                   string
	partial_data          []byte // whilst downloading this contains the partially downloaded bytes
	lock                  sync.Mutex
	completely_downloaded bool
	tobedeleted           bool
	readers               []*disk_reader
	abort_download        bool
}

func (dc *DiskCache) readonce() error {
	if dc.entries != nil {
		return nil
	}
	if !utils.FileExists(INDEXFILE) {
		dc.entries = make([]*diskCacheEntry, 0)
		return nil
	}
	fmt.Printf("Reading index file: %s\n", INDEXFILE)
	b, err := utils.ReadFile(INDEXFILE)
	if err != nil {
		return err
	}
	res := make([]*diskCacheEntry, 0)
	for _, s := range strings.Split(string(b), "\n") {
		if len(s) < 2 {
			continue
		}
		if s[0] == '#' {
			continue
		}
		de := parseDiskIndexEntry(s)
		if de != nil {
			if !de.completely_downloaded {
				de.tobedeleted = true
			}
			de.cache = dc
			res = append(res, de)
		}
	}
	dc.entries = res
	return nil
}
func parseDiskIndexEntry(s string) *diskCacheEntry {
	if len(s) < 2 {
		return nil
	}
	fs := strings.Split(s, " ")
	if len(fs) < 3 {
		fmt.Printf("Invalid cache line: \"%s\" (%d)\n", s, len(fs))
		return nil
	}
	ts, err := strconv.Atoi(fs[0])
	if err != nil {
		fmt.Printf("Invalid cache line: \"%s\", timestamp: %s\n", s, err)
		return nil
	}
	res := &diskCacheEntry{
		filename:     fs[1],
		url:          fs[2],
		lastModified: time.Unix(int64(ts), 0),
	}

	if len(fs) > 3 {
		ts, err = strconv.Atoi(fs[3])
		if err != nil {
			fmt.Printf("Invalid cache line: \"%s\", timestamp: %s\n", s, err)
			return nil
		}
		res.lastRead = time.Unix(int64(ts), 0)
	}

	if len(fs) > 4 {
		b := false
		if fs[4] == "true" {
			b = true
		}
		res.tobedeleted = b
	}

	if len(fs) > 5 {
		b := false
		if fs[5] == "true" {
			b = true
		}
		res.completely_downloaded = b
	}
	return res

}
func (dc *DiskCache) CacheSize() (uint64, error) {
	return 1, nil
}
func (dc *DiskCache) AddURL(url string) (CacheEntry, error) {
	dc.lock.Lock()
	defer dc.lock.Unlock()
	err := dc.readonce()
	if err != nil {
		return nil, err
	}
	ce := dc.findURLNoLock(url)
	if ce != nil {
		return nil, fmt.Errorf("already exists")
	}
	ce = &diskCacheEntry{
		cache:        dc,
		url:          url,
		lastModified: time.Now(),
		lastRead:     time.Now(),
		filename:     dc.findNextFreeFilename(),
	}
	dc.entries = append(dc.entries, ce)
	err = dc.flush()
	if err != nil {
		dc.entries = dc.entries[:len(dc.entries)-1]
		return nil, err
	}
	return ce, nil
}
func (dc *DiskCache) FindURL(url string) (CacheEntry, bool, error) {
	dc.lock.Lock()
	defer dc.lock.Unlock()
	err := dc.readonce()
	if err != nil {
		return nil, false, err
	}
	f := dc.findURLNoLock(url)
	if f == nil {
		return nil, false, nil
	}
	return f, true, nil
}
func (dc *DiskCache) findURLNoLock(url string) *diskCacheEntry {
	for _, ce := range dc.entries {
		if ce.url == url && ce.tobedeleted == false {
			return ce
		}
	}
	return nil
}

func (dc *DiskCache) findNextFreeFilename() string {
	i := 0
	for {
		i++
		fn := fmt.Sprintf("%s/%d", CACHEDIR, i)
		found := false
		for _, ce := range dc.entries {
			if ce.filename == fn {
				found = true
				break
			}
		}
		if !found {
			return fn
		}
	}
}

/***********************************************************
diskcache entry
***********************************************************/
func (ce *diskCacheEntry) DownloadFailure() {
	ce.tobedeleted = true
	ce.completely_downloaded = false
	ce.cache.request_flushing()

}
func (ce *diskCacheEntry) Size() uint64 {
	return 0
}
func (ce *diskCacheEntry) Reader() (io.ReadCloser, error) {
	ce.lastRead = time.Now()
	ce.cache.request_flushing()
	ds := &disk_reader{entry: ce, num: nextDiskReaderNum()}
	ce.addReader(ds)
	ds.read_from_mem = !ce.completely_downloaded
	return ds, nil
}

func (ce *diskCacheEntry) Write(b []byte) (int, error) {
	if ce.writefd == nil {
		dir := filepath.Dir(ce.filename)
		if !utils.FileExists(dir) {
			os.MkdirAll(dir, 0777)
		}
		fd, err := os.Create(ce.filename)
		if err != nil {
			return 0, err
		}
		ce.writefd = fd
	}
	ce.partial_data = append(ce.partial_data, b...)
	ce.lastModified = time.Now()
	ce.cache.request_flushing()
	return ce.writefd.Write(b)
}
func (ce *diskCacheEntry) WriteComplete() error {
	if ce.writefd == nil {
		return fmt.Errorf("file not open")
	}
	ce.completely_downloaded = true
	ce.lastModified = time.Now()
	ce.cache.request_flushing()
	err := ce.writefd.Close()
	return err
}
func (ce *diskCacheEntry) SerialiseToString() string {
	return fmt.Sprintf("%d %s %s %d %v %v",
		ce.lastModified.Unix(),
		ce.filename,
		ce.url,
		ce.lastRead.Unix(),
		ce.tobedeleted,
		ce.completely_downloaded,
	)
}
func (ce *diskCacheEntry) String() string {
	return fmt.Sprintf("%s->%s", ce.url, ce.filename)
}
func (ce *diskCacheEntry) MD5() (string, error) {
	f, err := os.Open(ce.filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	m := md5.New()
	if _, err := io.Copy(m, f); err != nil {
		return "", err
	}
	fmd5 := fmt.Sprintf("%x", m.Sum(nil))
	return fmd5, nil
}

func (ce *diskCacheEntry) MarkInUse() {
	ce.lastRead = time.Now()
	ce.cache.request_flushing()
}
func (ce *diskCacheEntry) IsDownloadCancelled() bool {
	return ce.abort_download
}
func (ce *diskCacheEntry) addReader(d *disk_reader) {
	ce.lock.Lock()
	ce.readers = append(ce.readers, d)
	ce.lock.Unlock()
}
func (ce *diskCacheEntry) removeReader(d *disk_reader) {
	ce.lock.Lock()
	defer ce.lock.Unlock()
	var x []*disk_reader
	for _, ed := range ce.readers {
		if ed == d {
			continue
		}
		x = append(x, ed)
	}
	if len(x) == len(ce.readers) {
		panic("no reader to remove!")
	}
	ce.readers = x

}
