package main

import (
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	"golang.conradwood.net/autodeployer/deployments"
	"golang.conradwood.net/autodeployer/downloader"
	chttp "golang.conradwood.net/go-easyops/http"
	"golang.conradwood.net/go-easyops/utils"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	idxname = "/srv/autodeployer/cache.index"
)

var (
	cache_latest  = flag.Bool("also_cache_latest", false, "if true it will also cache urls with the string 'latest' in. Normally it does not")
	cache_enabled = flag.Bool("cache_enable", true, "if false will not cache downloads")
	cache_v2      = flag.Bool("cache_v2", true, "if true, use v2 style caching")
	downloads     = make(map[string]*downloadlock)
	tr            = &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    120 * time.Second,
		DisableCompression: false,
	}

	downloadLock       sync.Mutex
	lastByteDownloaded time.Time
)

type downloadlock struct {
	lastUsed        time.Time
	lock            sync.Mutex
	downloading     bool
	downloadedBytes uint64
	totalBytes      uint64
	completed       bool
}

func cleanDownloadLocks() {
	if *cache_v2 {
		return
	}
	for k, v := range downloads {
		if time.Since(v.lastUsed) > time.Duration(30)*time.Minute {
			delete(downloads, k)
		}
	}
}

// this is called from deploymonkey to "pre-cache" a url
func (a *AutoDeployer) CacheURL(ctx context.Context, req *ad.URLRequest) (*ad.URLResponse, error) {
	fmt.Printf("Caching url %s\n", req.URL)
	if *cache_v2 {
		return downloader.CacheURL(ctx, req)
	}
	if !*cache_enabled {
		return &ad.URLResponse{URL: req.URL, BytesDownloaded: 0, TotalBytes: 0, CacheDisabled: true}, nil
	}
	url := req.URL
	downloadLock.Lock()
	cleanDownloadLocks()
	dlock := downloads[url]
	if dlock == nil {
		dlock = &downloadlock{}
		downloads[url] = dlock
	}
	dlock.lastUsed = time.Now()
	downloadLock.Unlock()
	if dlock.downloading || dlock.completed {
		return &ad.URLResponse{URL: req.URL, BytesDownloaded: dlock.downloadedBytes, TotalBytes: dlock.totalBytes}, nil
	}
	dlock.lock.Lock()
	defer func() {
		dlock.lock.Unlock()
	}()
	if req.ForceDownload {
		evictFromCache(url)
	}
	cf, ln, err := OpenCacheFileRead(url) // this will only return if it was successfully downloaded before
	if err == nil {
		cf.Close()
		return &ad.URLResponse{URL: req.URL, BytesDownloaded: ln, TotalBytes: ln}, nil
	}
	fmt.Printf("Failed to check cache file: %s (continuing without cache)\n", err)
	start_download(url)
	return &ad.URLResponse{URL: req.URL, BytesDownloaded: ln, TotalBytes: 1}, nil // just to make sure it is now considered complete yet
}

// this is called from the startup process and expects the binary
func (a *AutoDeployer) Download(req *ad.StartedRequest, srv ad.AutoDeployer_DownloadServer) error {
	d := entryByMsg(req.Msgid)
	if d == nil {
		return fmt.Errorf("No such startup message: \"%s\"", req.Msgid)
	}
	if *cache_v2 {
		d.Status = ad.DeploymentStatus_DOWNLOADING
		return downloader.Download(d.URL(), req, srv)
	}
	var cf *os.File
	var err error
	if *cache_enabled {
		cf, _, err = OpenCacheFileRead(d.URL())
		if err != nil {
			fmt.Printf("Failed to check cache file: %s (continuing without cache)\n", err)
		}
	}
	downloadLock.Lock()
	cleanDownloadLocks()
	dlock := downloads[d.URL()]
	if dlock == nil {
		dlock = &downloadlock{}
		downloads[d.URL()] = dlock
	}
	dlock.lastUsed = time.Now()
	downloadLock.Unlock()

	if req.DownloadFailed {
		evictFromCache(d.URL())
	}

	// get the lock and see if we now have a cache file
	dlock.lock.Lock()
	if *cache_enabled && cf == nil {
		// try again once we got the lock
		cf, _, err = OpenCacheFileRead(d.URL())
		if err != nil {
			fmt.Printf("Failed to check cache file: %s (continuing without cache)\n", err)
		}
	}
	dlock.lock.Unlock()

	if cf != nil {
		fmt.Printf("Serving %s from cache (%s)\n", d.URL(), cf.Name())
		defer cf.Close()
		d.Status = ad.DeploymentStatus_CACHEDSTART
		return ServeFromCache(d, cf, req, srv)
	}
	d.Status = ad.DeploymentStatus_DOWNLOADING

	dlock.lock.Lock()
	defer func() {
		dlock.downloading = false
		dlock.lock.Unlock()
	}()
	dlock.downloading = true

	fmt.Printf("Meant to download binary for \"%s\": %s\n", req.Msgid, d.URL())
	client := &http.Client{Transport: tr}

	hreq, err := http.NewRequest("GET", d.URL(), nil)
	if err != nil {
		fmt.Println("Error requesting", d.URL(), "-", err)
		return err
	}
	dr := d.DeployRequest
	if dr != nil && dr.DownloadUser != "" {
		hreq.SetBasicAuth(dr.DownloadUser, dr.DownloadPassword)
	}
	response, err := client.Do(hreq)
	if err != nil {
		fmt.Println("Error while downloading", d.URL(), "-", err)
		return err
	}
	defer response.Body.Close()
	code := response.StatusCode
	if code < 200 || code > 299 {
		return fmt.Errorf("download %s failed with code %d", d.URL(), code)
	}
	cf = nil
	if *cache_enabled {
		if *cache_latest || !strings.Contains(d.URL(), "latest") {
			lastByteDownloaded = time.Now()
			cf, err = CreateCacheFile()
			if err != nil {
				fmt.Printf("Failed to open cache file (for writing): %s\n", err)
			}
			fmt.Printf("Writing to cache file: %s\n", cf.Name())
		}
	}
	buf := make([]byte, 32000)
	var forward_error error
	pr := utils.ProgressReporter{}
	size := get_size_from_response(response)
	dlock.totalBytes = size
	pr.SetTotal(size)
	for {
		lastByteDownloaded = time.Now()
		n, err := response.Body.Read(buf)
		pr.Add(uint64(n))
		pr.Print()
		sb := buf[:n]
		if n > 0 {
			bd := &ad.BinaryDownload{Data: sb}
			if forward_error == nil {
				serr := srv.Send(bd)
				if serr != nil {
					forward_error = serr
				}
			}
			if cf != nil {
				_, err = cf.Write(sb)
				if err != nil {
					fmt.Printf("failed to write to cache file: %s\n", err)
					cf.Close()
					return err
				}
			}

		}
		if err == io.EOF {
			break
		}
		if err != nil {
			cf.Close()
			return err
		}

	}
	if cf != nil {
		err = cf.Close()
		if err != nil {
			return err
		}
		err := checkMD5(cf, d)
		if err != nil {
			// file is corrupt - remove it
			os.Remove(cf.Name())
			return err
		}
	}
	if *cache_enabled && cf != nil {
		err = MarkCache(cf, d.URL())
		if err != nil {
			return err
		}
		dlock.completed = true

	}
	if forward_error != nil {
		return forward_error
	}
	return nil
}

func MarkCache(file *os.File, url string) error {
	if *cache_v2 {
		return nil
	}

	if file == nil {
		return nil
	}
	downloadLock.Lock()
	defer downloadLock.Unlock()
	s := ""
	if utils.FileExists(idxname) {
		ci, err := utils.ReadFile(idxname)
		if err != nil {
			return err
		}
		s = string(ci)
	}
	now := time.Now().Unix()
	// strip out existing caches for this url
	news := ""
	for _, l := range strings.Split(s, "\n") {
		if len(l) < 2 {
			continue
		}
		if strings.HasSuffix(l, url) {
			continue
		}
		news = news + l + "\n"
	}
	news = news + fmt.Sprintf("%d %s %s\n", now, file.Name(), url)
	err := utils.WriteFile(idxname, []byte(news))
	if err != nil {
		fmt.Printf("Failed to write index: %s\n", err)
	} else {
		fmt.Printf("Added %s to cache\n", url)
	}
	return err
}

func CreateCacheFile() (*os.File, error) {
	if *cache_v2 {
		panic("v1 cache should be disabled!")
	}

	dir := "/srv/autodeployer/download_cache"
	os.MkdirAll(dir, 0777)
	downloadLock.Lock()
	defer downloadLock.Unlock()
	i := 1
	for {
		fname := fmt.Sprintf("%s/%d", dir, i)
		if utils.FileExists(fname) {
			i++
			continue
		}
		cf, err := os.Create(fname)
		if err != nil {
			return nil, err
		}
		return cf, nil
	}

}

// open cache file for url for READ (also return number of bytes)
func OpenCacheFileRead(url string) (*os.File, uint64, error) {
	if *cache_v2 {
		panic("v1 cache should be disabled!")
	}

	filename, err := GetCacheFilename(url)
	if err != nil {
		return nil, 0, err
	}
	if filename == "" {
		return nil, 0, fmt.Errorf("unable to determine filename for url %s\n", url)
	}
	st, err := os.Stat(filename)
	if err != nil {
		return nil, 0, fmt.Errorf("Error stating cachefile for url \"%s\" (%s): %s", url, filename, err)
	}
	ln := uint64(st.Size())
	fop, err := os.Open(filename)
	if err != nil {
		return nil, ln, fmt.Errorf("error opening cache for url \"%s\" (%s): %s", url, filename, err)
	}
	return fop, ln, err
}
func isDownloading(url string) bool {
	if *cache_v2 {
		panic("v1 cache should be disabled!")
	}

	downloadLock.Lock()
	defer downloadLock.Unlock()
	dl := downloads[url]
	if dl == nil {
		return false
	}
	return dl.downloading

}
func evictFromCache(url string) {
	if *cache_v2 {
		panic("v1 cache should be disabled!")
	}

	if isDownloading(url) {
		fmt.Printf("Not evicting url %s - it is still downloading\n", url)
		// whilst downloading it is bound to be broken
		return
	}
	filename, err := GetCacheFilename(url)
	if err != nil {
		fmt.Printf("Failed to evict: %s\n", err)
		return
	}
	err = os.Remove(filename)
	if err != nil {
		fmt.Printf("failed to remove file: %s\n", err)
		return
	}
	fmt.Printf("%s evicted from cache.\n", url)
}

// return filename of cache file (or "")
func GetCacheFilename(url string) (string, error) {
	if *cache_v2 {
		panic("v1 cache should be disabled!")
	}

	if !utils.FileExists(idxname) {
		return "", nil
	}
	ci, err := utils.ReadFile(idxname)
	if err != nil {
		return "", err
	}
	s := string(ci)
	for i, l := range strings.Split(s, "\n") {
		if l == "" {
			continue
		}
		fields := strings.SplitN(l, " ", 3)
		if len(fields) != 3 {
			fmt.Printf("Invalid line %d in index file: \"%s\"\n", i, l)
			continue
		}
		filename := fields[1]
		curl := fields[2]
		//		fmt.Printf("URL \"%s\" in file \"%s\"\n", curl, filename)
		if url == curl {
			return filename, nil
		}
	}
	return "", nil
}

func ServeFromCache(d *deployments.Deployed, cf *os.File, req *ad.StartedRequest, srv ad.AutoDeployer_DownloadServer) error {
	buf := make([]byte, 8192)
	for {
		n, err := cf.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			cf.Close()
			return err
		}
		sb := buf[:n]
		bd := &ad.BinaryDownload{Data: sb}
		err = srv.Send(bd)
		if err != nil {
			cf.Close()
			return err
		}
	}
	err := cf.Close()
	if err != nil {
		return err
	}
	return nil
}
func start_download(url string) {
	if *cache_v2 {
		panic("v1 cache should be disabled!")
	}

	go func(u string) {
		err := download(u)
		if err != nil {
			fmt.Printf("Error downloading \"%s\": %s\n", u, err)
		}
	}(url)
}

func download(url string) error {
	if *cache_v2 {
		panic("v1 cache should be disabled!")
	}

	if !*cache_enabled {
		return fmt.Errorf("cache disabled")
	}
	if !*cache_latest && strings.Contains(url, "latest") {
		return fmt.Errorf("not caching url containing the word 'latest'")
	}
	dlock := downloads[url]
	if dlock != nil && dlock.downloading {
		// downloading already..
		return nil
	}
	downloadLock.Lock()
	cleanDownloadLocks()
	dlock = downloads[url]
	if dlock == nil {
		dlock = &downloadlock{}
		downloads[url] = dlock
	}
	dlock.lastUsed = time.Now()
	downloadLock.Unlock()
	dlock.lock.Lock()
	if dlock.downloading {
		// downloading already..
		dlock.lock.Unlock()
		return nil
	}
	dlock.downloading = true
	defer func() {
		dlock.downloading = false
		dlock.lock.Unlock()
	}()
	client := &http.Client{Transport: tr}

	hreq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error requesting", url, "-", err)
		return err
	}
	response, err := client.Do(hreq)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return err
	}
	defer response.Body.Close()
	code := response.StatusCode
	if code < 200 || code > 299 {
		return fmt.Errorf("download %s failed with code %d", url, code)
	}

	cf, err := CreateCacheFile()
	if err != nil {
		return fmt.Errorf("Failed to open cache file (for writing): %s\n", err)

	}
	fmt.Printf("Writing to cache file: %s\n", cf.Name())

	buf := make([]byte, 32000)
	dlock.downloadedBytes = 0
	pr := utils.ProgressReporter{}
	size := get_size_from_response(response)
	dlock.totalBytes = size
	pr.SetTotal(size)
	for {
		n, err := response.Body.Read(buf)
		pr.Add(uint64(n))
		pr.Print()
		dlock.downloadedBytes = dlock.downloadedBytes + uint64(n)
		sb := buf[:n]
		if n > 0 {
			_, err = cf.Write(sb)
			if err != nil {
				fmt.Printf("failed to write to cache file: %s\n", err)
				cf.Close()
				return err
			}

		}
		if err == io.EOF {
			break
		}
		if err != nil {
			cf.Close()
			return err
		}
	}
	cf.Close()
	err = MarkCache(cf, url)
	if err == nil {
		dlock.completed = true
	}
	return err
}

func get_size_from_response(r *http.Response) uint64 {
	var err error
	size := uint64(0)
	for k, v := range r.Header {
		if strings.ToLower(k) == "content-length" {
			if len(v) > 0 {
				size, err = strconv.ParseUint(v[0], 10, 64)
				if err != nil {
					fmt.Printf("Invalid content length \"%s\": %s", v, err)
					return 0
				}
			}
		}
	}
	return size
}

func checkMD5(cf *os.File, d *deployments.Deployed) error {
	fname := cf.Name()
	md5url := d.URL() + ".md5"
	fmt.Printf("Checking md5 on file \"%s\" from \"%s\"\n", fname, md5url)
	h := chttp.HTTP{}
	hr := h.Get(md5url)
	err := hr.Error()
	if err != nil {
		fmt.Printf("unable to download md5: %s\n", err)
		// does not exist - not an error, we just skip the test
		return nil
	}
	s := string(hr.Body())
	fs := strings.Fields(s)
	if len(fs) < 1 {
		fmt.Printf("Invalid md5 body (no fields): %s\n", s)
		return nil
	}
	md5s := fs[0]
	fmt.Printf("Downloaded md5: \"%s\", compare with %s\n", md5s, fname)

	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	m := md5.New()
	if _, err := io.Copy(m, f); err != nil {
		return err
	}
	fmd5 := fmt.Sprintf("%x", m.Sum(nil))
	if md5s != fmd5 {
		return fmt.Errorf("Corrupted download (md5sum mismatch)")
	}

	return nil
}
