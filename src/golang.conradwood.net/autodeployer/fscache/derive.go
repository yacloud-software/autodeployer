package fscache

import (
	"fmt"
	pb "golang.conradwood.net/apis/commondeploy"
	"golang.conradwood.net/go-easyops/utils"
	"io"
	"os"
	"time"
)

type derive_function struct {
	df func(io.Reader, io.Writer) error
}

func (f *fscache) RegisterDeriveFunction(function_id string, ff func(io.Reader, io.Writer) error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.derive_functions[function_id] = &derive_function{df: ff}
}

func (f *fscache) GetDerivedFile(ce *pb.CacheEntry, file_id string, funcname string) ([]byte, error) {
	f.lock.Lock()
	ce, err := f.reloadEntry(ce)
	if err != nil {
		f.lock.Unlock()
		return nil, err
	}
	f.debugf("Deriving in cachentry %#v\n", ce)
	var dce *pb.DerivedCacheEntry
	for _, de := range ce.DerivedEntries {
		if de.FileID == file_id {
			dce = de
			if dce.Completed {
				f.lock.Unlock()
				f.debugf("Derived file %s served from cache", file_id)
				return f.read_bytes_for_derived(ce, dce)
			}
			if dce.Deriving {
				f.debugf("Derived file %s in progress atm", file_id)
				f.lock.Unlock()
				return f.wait_for_derive(ce, dce)
			}
			break
		}
	}
	if dce != nil && dce.Deriving {
		f.lock.Unlock()
		return f.wait_for_derive(ce, dce)
	}
	f.debugf("Deriving %s from %#v", file_id, dce)
	// no such file - must derive
	ff, found := f.derive_functions[funcname]
	if !found {
		return nil, fmt.Errorf("no such function \"%s\"", funcname)
	}

	if dce == nil {
		dce = &pb.DerivedCacheEntry{
			FileID:    file_id,
			Function:  funcname,
			Completed: false,
			FileRef:   utils.RandomString(64),
		}
		ce.DerivedEntries = append(ce.DerivedEntries, dce)
	}
	dce.Deriving = true
	err = f.updateEntry(ce)
	if err != nil {
		f.lock.Unlock()
		return nil, err
	}

	f.lock.Unlock()

	r, err := os.Open(f.get_cache_dir(ce) + "/orig_file")
	if err != nil {
		return nil, err
	}
	w, err := os.Create(f.get_cache_dir(ce) + "/" + dce.FileRef)
	if err != nil {
		return nil, err
	}
	err = ff.df(r, w) // do the actual conversion
	if err != nil {
		f.debugf("Derive failed (%s)", err)
		f.lock.Lock()
		dce.Deriving = false
		f.updateDerived(ce, dce)
		f.lock.Unlock()
		return nil, err
	}
	f.debugf("Derive complete")
	f.lock.Lock()
	dce.Deriving = false
	dce.Completed = true
	xerr := f.updateDerived(ce, dce)
	f.lock.Unlock()
	if xerr != nil {
		return nil, xerr
	}
	return nil, nil
}

func (f *fscache) wait_for_derive(ce *pb.CacheEntry, dce *pb.DerivedCacheEntry) ([]byte, error) {
	for {
		ce, err := f.reloadEntry(ce)
		if err != nil {
			return nil, err
		}
		var nd *pb.DerivedCacheEntry
		for _, xdce := range ce.DerivedEntries {
			if dce.FileID == xdce.FileID {
				nd = xdce
				break
			}
		}
		if nd == nil {
			return nil, fmt.Errorf("no such derived cache entry")
		}
		if nd.Completed {
			return f.read_bytes_for_derived(ce, nd)
		}
		if !nd.Deriving {
			return nil, fmt.Errorf("not deriving, cannot wait for it")
		}
		time.Sleep(time.Duration(1) * time.Second)
	}

}
func (f *fscache) read_bytes_for_derived(ce *pb.CacheEntry, dce *pb.DerivedCacheEntry) ([]byte, error) {
	fname := f.get_cache_dir(ce) + "/" + dce.FileRef
	res, err := utils.ReadFile(fname)
	return res, err
}
func (f *fscache) updateDerived(ce *pb.CacheEntry, dce *pb.DerivedCacheEntry) error {
	ce, err := f.reloadEntry(ce)
	if err != nil {
		return err
	}
	for i, xdce := range ce.DerivedEntries {
		if dce.FileID == xdce.FileID {
			ce.DerivedEntries[i] = dce
			return f.updateEntry(ce)
		}
	}
	return fmt.Errorf("no such derived entry")
}
