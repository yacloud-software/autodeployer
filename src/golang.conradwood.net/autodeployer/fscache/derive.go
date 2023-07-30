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
	isFile bool
	df     func(io.Reader, io.Writer) error // derive a new file (e.g. unzip)
	dd     func(io.Reader, string) error    // derive a dir (e.g. untar)
}

func (f *fscache) RegisterDeriveFunction(function_id string, ff func(io.Reader, io.Writer) error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.derive_functions[function_id] = &derive_function{df: ff, isFile: true}
}
func (f *fscache) RegisterDeriveFunctionDir(function_id string, ff func(io.Reader, string) error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.derive_functions[function_id] = &derive_function{dd: ff, isFile: false}
}

func (f *fscache) GetDerivedFile(ce *pb.CacheEntry, file_id string, funcname string) (string, error) {
	f.lock.Lock()
	ce, err := f.reloadEntry(ce)
	if err != nil {
		f.lock.Unlock()
		return "", err
	}
	f.debugf("Deriving in cachentry %#v\n", ce)
	var dce *pb.DerivedCacheEntry
	for _, de := range ce.DerivedEntries {
		if de.FileID == file_id {
			dce = de
			if dce.Completed {
				f.lock.Unlock()
				f.debugf("Derived file %s served from cache", file_id)
				return f.get_filename_for_derived(ce, dce)
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
		return "", fmt.Errorf("no such function \"%s\"", funcname)
	}

	if dce == nil {
		dce = &pb.DerivedCacheEntry{
			FileID:    file_id,
			Function:  funcname,
			Completed: false,
			FileRef:   utils.RandomString(64),
			LastUsed:  uint32(time.Now().Unix()),
		}
		ce.DerivedEntries = append(ce.DerivedEntries, dce)
	}
	dce.Deriving = true
	err = f.updateEntry(ce)
	if err != nil {
		f.lock.Unlock()
		return "", err
	}

	f.lock.Unlock()

	r, err := os.Open(f.get_cache_dir(ce) + "/orig_file")
	if err != nil {
		return "", err
	}
	defer r.Close()

	// difference between file (e.g. unzip) and creating a dir (e.g. untar)
	if ff.isFile {
		w, err := os.Create(f.get_cache_dir(ce) + "/" + dce.FileRef)
		if err != nil {
			return "", err
		}
		err = ff.df(r, w) // do the actual conversion
		w.Close()
	} else {
		panic("cannot derive dir yet")
	}

	if err != nil {
		f.debugf("Derive failed (%s)", err)
		f.lock.Lock()
		dce.Deriving = false
		f.updateDerived(ce, dce)
		f.lock.Unlock()
		return "", err
	}
	f.debugf("Derive complete")
	f.lock.Lock()
	dce.Deriving = false
	dce.Completed = true
	xerr := f.updateDerived(ce, dce)
	f.lock.Unlock()
	if xerr != nil {
		return "", xerr
	}
	return "", nil
}

func (f *fscache) GetDerivedFileFromDerived(ce *pb.CacheEntry, from_id, file_id string, function_id string) (string, error) {
	f.lock.Lock()
	ce, err := f.reloadEntry(ce)
	if err != nil {
		f.lock.Unlock()
		return "", err
	}

	// first check if it's already done
	var tce *pb.DerivedCacheEntry
	for _, de := range ce.DerivedEntries {
		if de.FileID == file_id {
			tce = de
			if de.Completed {
				f.lock.Unlock()
				f.debugf("serving derived \"%s\" (from \"%s\") from cache", file_id, from_id)
				return f.get_filename_for_derived(ce, de)
			}
		}
	}
	if tce != nil && tce.Deriving {
		f.lock.Unlock()
		return f.wait_for_derive(ce, tce)
	}

	// we have to build it, so check if we got the function

	ff, found := f.derive_functions[function_id]
	if !found {
		return "", fmt.Errorf("no such function \"%s\"", function_id)
	}

	// we have to build it, so check if the derived one is ready
	f.debugf("Deriving in cachentry %#v\n", ce)
	var dce *pb.DerivedCacheEntry
	for _, de := range ce.DerivedEntries {
		if de.FileID == from_id {
			dce = de
			break
		}
	}
	if dce == nil {
		f.lock.Unlock()
		return "", fmt.Errorf("no derived file \"%s\"", from_id)
	}
	if dce.Deriving {
		f.lock.Unlock()
		_, err = f.wait_for_derive(ce, dce)
		if err != nil {
			return "", err
		}
	}
	// the file from which we are deriving is ready
	if tce == nil {
		// must create a new derived entry
		tce = &pb.DerivedCacheEntry{
			FileID:    file_id,
			Function:  function_id,
			Completed: false,
			FileRef:   utils.RandomString(64),
			LastUsed:  uint32(time.Now().Unix()),
			Deriving:  true,
		}
		ce.DerivedEntries = append(ce.DerivedEntries, tce)
	}
	dce.Deriving = true
	err = f.updateEntry(ce)
	if err != nil {
		f.lock.Unlock()
		return "", err
	}

	// actually create it...
	sourcefile, err := f.get_filename_for_derived(ce, dce)
	f.lock.Unlock()
	if err != nil {
		return "", err
	}
	r, err := os.Open(sourcefile)
	if err != nil {
		return "", err
	}
	defer r.Close()

	// difference between file (e.g. unzip) and creating a dir (e.g. untar)
	target := f.get_cache_dir(ce) + "/" + tce.FileRef
	if ff.isFile {
		w, err := os.Create(target)
		if err != nil {
			return "", err
		}
		err = ff.df(r, w) // do the actual conversion
		w.Close()
	} else {
		err = ff.dd(r, target) // do the actual conversion
	}
	if err != nil {
		f.debugf("Derive failed (%s)", err)
		f.lock.Lock()
		tce.Deriving = false
		f.updateDerived(ce, tce)
		f.lock.Unlock()
		return "", err
	}
	f.debugf("Derive complete")
	f.lock.Lock()
	tce.Deriving = false
	tce.Completed = true
	xerr := f.updateDerived(ce, tce)
	f.lock.Unlock()
	if xerr != nil {
		return "", xerr
	}

	// the derived file is ready

	return target, nil
}

func (f *fscache) wait_for_derive(ce *pb.CacheEntry, dce *pb.DerivedCacheEntry) (string, error) {
	for {
		ce, err := f.reloadEntry(ce)
		if err != nil {
			return "", err
		}
		var nd *pb.DerivedCacheEntry
		for _, xdce := range ce.DerivedEntries {
			if dce.FileID == xdce.FileID {
				nd = xdce
				break
			}
		}
		if nd == nil {
			return "", fmt.Errorf("no such derived cache entry")
		}
		if nd.Completed {
			return f.get_filename_for_derived(ce, nd)
		}
		if !nd.Deriving {
			return "", fmt.Errorf("not deriving, cannot wait for it")
		}
		time.Sleep(time.Duration(1) * time.Second)
	}

}
func (f *fscache) read_bytes_for_derived(ce *pb.CacheEntry, dce *pb.DerivedCacheEntry) ([]byte, error) {
	fname := f.get_cache_dir(ce) + "/" + dce.FileRef
	res, err := utils.ReadFile(fname)
	return res, err
}
func (f *fscache) get_filename_for_derived(ce *pb.CacheEntry, dce *pb.DerivedCacheEntry) (string, error) {
	dce.LastUsed = uint32(time.Now().Unix())
	f.updateDerived(ce, dce)
	fname := f.get_cache_dir(ce) + "/" + dce.FileRef
	return fname, nil
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
