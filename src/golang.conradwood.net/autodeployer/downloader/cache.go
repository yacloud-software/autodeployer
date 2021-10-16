package downloader

import (
	"io"
)

type Cache interface {
	CacheSize() (uint64, error)                   // size of cache
	AddURL(url string) (CacheEntry, error)        // create a new file (overwrite if exists)
	FindURL(url string) (CacheEntry, bool, error) // find a cacheentry by url , true if exists
}

type CacheEntry interface {
	MarkInUse()
	IsDownloadCancelled() bool
	Size() uint64
	Reader() (io.ReadCloser, error)
	Write(p []byte) (n int, err error)
	WriteComplete() error // called when file is completely downloaded
	DownloadFailure()     // call if file is partially downloaded and failed
	String() string
	MD5() (string, error)
}
