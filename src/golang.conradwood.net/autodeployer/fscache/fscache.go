package fscache

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/commondeploy"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/utils"
	"io"
	"os"
	"sync"
	"time"
)

type FSCache interface {
	// retrieve and download if necessary
	Cache(ctx context.Context, cr *pb.CacheRequest) (*pb.CacheEntry, error)
	// retrieve, do not download, return error if not in cache
	ReadCache(ctx context.Context, cr *pb.CacheRequest) (*pb.CacheEntry, error)
	// currently active downloads for this instance
	ActiveDownloads() int
	// user may register functions to derive files from others, e.g. "de-bzip2"
	RegisterDeriveFunction(function_id string, f func(io.Reader, io.Writer) error)
	// derive another file from the (fully downloaded) cache entry and cache it for future (returns fully qualified filename)
	GetDerivedFile(ce *pb.CacheEntry, file_id string, function_id string) (string, error)
}

type fscache struct {
	maxsizeMB        int
	lock             sync.Mutex
	statedir         string
	maxdownloads     int // maximum simultaneous downloads
	active_downloads int // current downloads
	d_lock           sync.Mutex
	derive_functions map[string]*derive_function
}

func NewFSCache(maxsizeMB int, statedir string) FSCache {
	res := &fscache{
		maxsizeMB:        maxsizeMB,
		statedir:         statedir,
		maxdownloads:     4,
		derive_functions: make(map[string]*derive_function),
	}
	// register default functions
	register_default_functions(res)
	res.resetDownloads()

	return res
}
func (f *fscache) ActiveDownloads() int {
	return f.active_downloads
}
func (f *fscache) resetDownloads() error {
	f.lock.Lock()
	defer f.lock.Unlock()
	cache_list_file := f.statedir + "/cache_list_file"
	cl, err := read_cache_list_file(cache_list_file)
	if err != nil {
		return err
	}
	for _, c := range cl.entries {
		c.Downloading = false
		for _, dce := range c.DerivedEntries {
			dce.Deriving = false
		}
	}
	err = cl.write()
	if err != nil {
		return err
	}
	return nil
}

/*
get a file from cache, do not retrieve it if it does not exist
*/
func (f *fscache) ReadCache(ctx context.Context, cr *pb.CacheRequest) (*pb.CacheEntry, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	cache_entry, err := f.findEntry(ctx, cr)
	if err != nil {
		return nil, err
	}
	if cache_entry == nil {
		return nil, errors.NotFound(ctx, "cache entry %s not found", cr.URL)
	}
	return cache_entry, nil
}

/*
get a file from cache, retrieve it if necessary
*/
func (f *fscache) Cache(ctx context.Context, cr *pb.CacheRequest) (*pb.CacheEntry, error) {
	f.lock.Lock()
	cache_entry, err := f.findEntry(ctx, cr)
	if err != nil {
		f.lock.Unlock()
		return nil, err
	}
	if cache_entry != nil {
		f.lock.Unlock()
		if cache_entry.Downloaded == true {
			f.debugf("returning cache entry for %v", cr)
			return cache_entry, nil
		}
		ce, err := f.handleIncompleteCacheEntry(ctx, cache_entry)
		if err != nil {
			return nil, err
		}
		cache_entry = ce
		if cache_entry == nil {
			panic("empty cache entry")
		}
		f.debugf("Found (completed) cache entry %v for %s\n", cache_entry, cr)
		return cache_entry, nil

	} else {
		// must download it...
		cd := f.uniqueDir()
		cache_entry = &pb.CacheEntry{
			CachedURL:       cr.URL,
			Downloading:     true,
			DownloadStarted: uint32(time.Now().Unix()),
			CacheDir:        cd,
		}
		os.MkdirAll(f.get_cache_dir(cache_entry), 0777)
		f.addEntry(ctx, cache_entry)
	}
	f.lock.Unlock()

	err = f.download(cache_entry)
	f.lock.Lock()
	defer f.lock.Unlock()
	if err == nil {
		cache_entry.Downloaded = true
		f.updateEntry(cache_entry)
		return cache_entry, nil
	}
	cache_entry.Failures++
	cache_entry.Downloaded = false
	cache_entry.LastError = fmt.Sprintf("%s", err)
	cache_entry.Downloading = false
	f.updateEntry(cache_entry)
	return nil, err
}

func (f *fscache) handleIncompleteCacheEntry(ctx context.Context, ce *pb.CacheEntry) (*pb.CacheEntry, error) {
	f.debugf("handling incomplete entry %v", ce)
	for {
		f.lock.Lock()
		cache_list_file := f.statedir + "/cache_list_file"
		cl, err := read_cache_list_file(cache_list_file)
		if err != nil {
			f.lock.Unlock()
			return nil, err
		}
		var cache_entry *pb.CacheEntry
		for _, xce := range cl.entries {
			if xce.CachedURL == ce.CachedURL {
				cache_entry = xce
				break
			}
		}

		if cache_entry.Downloaded {
			f.lock.Unlock()
			return cache_entry, nil
		}

		if !cache_entry.Downloading {
			cache_entry.Downloading = true
			cache_entry.DownloadStarted = uint32(time.Now().Unix())
			err = f.updateEntry(cache_entry)
			f.lock.Unlock()
			if err != nil {
				return nil, err
			}
			err = f.download(cache_entry)
			f.lock.Lock()
			cache_entry.Downloading = false
			if err != nil {
				cache_entry.LastError = fmt.Sprintf("%s", err)
				cache_entry.Failures++
				f.updateEntry(cache_entry)
				f.lock.Unlock()
				return nil, err
			}
			cache_entry.Downloaded = true
			cache_entry.DownloadedTimestamp = uint32(time.Now().Unix())
			err = f.updateEntry(cache_entry)
			if err != nil {
				return nil, err
			}
			f.lock.Unlock()
			return cache_entry, nil
		}

		f.lock.Unlock()
		time.Sleep(time.Duration(3) * time.Second)

	}
}

func (f *fscache) debugf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	fmt.Println("[fscache] " + s)
}
func (f *fscache) reloadEntry(ce *pb.CacheEntry) (*pb.CacheEntry, error) {
	cache_list_file := f.statedir + "/cache_list_file"
	cl, err := read_cache_list_file(cache_list_file)
	if err != nil {
		return nil, err
	}
	for _, xce := range cl.entries {
		if xce.CachedURL == ce.CachedURL {
			return xce, nil
		}
	}
	return nil, fmt.Errorf("cacheentry does not exist")
}
func (f *fscache) updateEntry(ce *pb.CacheEntry) error {
	cache_list_file := f.statedir + "/cache_list_file"
	cl, err := read_cache_list_file(cache_list_file)
	if err != nil {
		return err
	}
	for i, xce := range cl.entries {
		if xce.CachedURL == ce.CachedURL {
			cl.entries[i] = ce
		}
	}
	err = cl.write()
	return err
}

func (f *fscache) addEntry(ctx context.Context, ce *pb.CacheEntry) error {
	cache_list_file := f.statedir + "/cache_list_file"
	cl, err := read_cache_list_file(cache_list_file)
	if err != nil {
		return err
	}
	cl.add(ce)
	err = cl.write()
	if err != nil {
		return err
	}
	return nil
}

// find an entry, return nil if non existent on disk
func (f *fscache) findEntry(ctx context.Context, cr *pb.CacheRequest) (*pb.CacheEntry, error) {
	cache_list_file := f.statedir + "/cache_list_file"
	cl, err := read_cache_list_file(cache_list_file)
	if err != nil {
		return nil, err
	}
	for _, ce := range cl.entries {
		if ce.CachedURL == cr.URL {
			return ce, nil
		}
	}
	return nil, nil
}

func (f *fscache) uniqueDir() string {
	i := 0
	for {
		i++
		sfname := fmt.Sprintf("%06d", i)
		fname := fmt.Sprintf("%s/cachedir/%s", f.statedir, sfname)
		if !utils.FileExists(fname) {
			f.debugf("new directory: %s (%s)", fname, sfname)
			return sfname
		}
	}
}
func (f *fscache) get_cache_dir(ce *pb.CacheEntry) string {
	return f.statedir + "/cachedir/" + ce.CacheDir
}
