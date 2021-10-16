package downloader

import (
	"context"
	"flag"
	ad "golang.conradwood.net/apis/autodeployer"
	"sync"
	"testing"
)

var (
	onc sync.Once
)

func test_setup() {
	flag.Parse()
	Start(func() bool {
		return true
	})
}
func TestCache(t *testing.T) {
	onc.Do(test_setup)
	t.Logf("testing...\n")
	dc := NewDiskCache()
	dc.entries = make([]*diskCacheEntry, 0)
	createAndCheck(t, dc, "foo")
	createAndCheck(t, dc, "bar")

	url := "https://www.conradwood.net/downloads/testfile.bin"
	_, err := CacheURL(context.Background(), &ad.URLRequest{URL: url})
	if err != nil {
		t.Fatalf("failed: %s\n", err)
	}
	sr := &ad.StartedRequest{}
	err = Download(url, sr, nil)
	if err != nil {
		t.Fatalf("download failed: %s", err)
	}

}

func createAndCheck(t *testing.T, dc *DiskCache, url string) {
	ol := len(dc.entries)
	_, err := dc.AddURL(url)
	if err != nil {
		t.Fatalf("failed to add url: %s", err)
	}
	if len(dc.entries) != ol+1 {
		t.Fatalf("dc.entries did not increase by 1 (expected from %d to %d, but is %d)", ol, ol+1, len(dc.entries))
	}
	_, found, err := dc.FindURL(url)
	if err != nil {
		t.Fatalf("failed to find url: %s", err)
	}
	if !found {
		t.Fatalf("did not find url (%s)", url)
	}
}
