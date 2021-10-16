package downloader

import (
	"context"
	"flag"
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	gh "golang.conradwood.net/go-easyops/http"
	"golang.conradwood.net/go-easyops/prometheus"
	"golang.conradwood.net/go-easyops/utils"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	download_stall = flag.Float64("download_stall", 0, "number in seconds between chunks to stall download")
	tr             = &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    120 * time.Second,
		DisableCompression: false,
	}

	cache     *DiskCache
	isEnabled func() bool
)

// do not use init here - because init() is also called for the startup code
func Start(f func() bool) {
	prometheus.MustRegister(readersGauge, partialGauge)
	isEnabled = f
	cache = NewDiskCache()
	go flush_thread()
}

// a url is currently in use (mark as lastread)
func URLActive(url string) {
	//fmt.Printf("Active: \"%s\"\n", url)
	dc, found, err := cache.FindURL(url)
	if err != nil || !found || dc == nil {
		fmt.Printf("URL active, but not cached: \"%s\"\n", url)
		return
	}
	dc.MarkInUse()

}

func CacheURL(ctx context.Context, req *ad.URLRequest) (*ad.URLResponse, error) {
	fmt.Printf("Request to cache url %s\n", req.URL)
	dc, found, err := cache.FindURL(req.URL)
	if err != nil {
		return nil, err
	}
	if found {
		fmt.Printf("URL %s is cached (%s)\n", req.URL, dc.String())
		return &ad.URLResponse{
			URL:             req.URL,
			BytesDownloaded: dc.Size(),
			TotalBytes:      dc.Size(),
			CacheDisabled:   false,
		}, nil
	}
	dc, err = cache.AddURL(req.URL)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Starting download for %s\n", dc.String())
	err = start_download(req.URL, dc)
	if err != nil {
		// evict?
		return nil, err
	}
	return &ad.URLResponse{
		URL:             req.URL,
		BytesDownloaded: 0,
		TotalBytes:      dc.Size(),
		CacheDisabled:   false,
	}, nil

}
func Download(url string, req *ad.StartedRequest, srv ad.AutoDeployer_DownloadServer) error {
	fmt.Printf("Request to download url %s\n", url)
	dc, found, err := cache.FindURL(url)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("URL %s is not cached yet", url)
	}
	fmt.Printf("cached at: %s\n", dc.String())
	buf := make([]byte, 32768)
	reader, err := dc.Reader()
	if err != nil {
		return err
	}
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			bd := &ad.BinaryDownload{Data: buf[:n]}
			if srv != nil {
				srv.Send(bd)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			reader.Close()
			return err
		}

	}
	err = reader.Close()
	if err != nil {
		return err
	}
	return nil
}

/**********************************************************************************************
* http download stuff
**********************************************************************************************/

func start_download(url string, ce CacheEntry) error {
	fmt.Printf("Downloading url %s\n", url)
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
	code := response.StatusCode
	if code < 200 || code > 299 {
		return fmt.Errorf("download %s failed with code %d", url, code)
	}
	go download_body(url, ce, response)
	return nil
}

// this downloads the body. this might take some time..
func download_body(url string, ce CacheEntry, response *http.Response) error {
	defer response.Body.Close()
	buf := make([]byte, 32000)
	p := utils.ProgressReporter{Prefix: url}
	if response.ContentLength != -1 {
		p.SetTotal(uint64(response.ContentLength))
	}
	for {
		if *download_stall != 0 {
			time.Sleep(time.Duration(*download_stall) * time.Second)
		}
		n, err := response.Body.Read(buf)
		p.Add(uint64(n))
		p.Print()
		if n > 0 {
			n2, err := ce.Write(buf[:n])
			if err != nil {
				fmt.Printf("Write error: %s\n", err)
				ce.DownloadFailure()
				return err
			}
			if n2 != n {
				fmt.Printf("incomplete write error: %d!=%d\n", n2, n)
				ce.DownloadFailure()
				return fmt.Errorf("Instead of %d bytes, only wrote %d\n", n, n2)
			}
		}
		if ce.IsDownloadCancelled() {
			fmt.Printf("Download aborted.\n")
			ce.DownloadFailure()
			return fmt.Errorf("download aborted")
		}
		if err == io.EOF {
			break
		}
	}
	err := ce.WriteComplete()
	if err != nil {
		fmt.Printf("write not completed: %s\n", err)
		ce.DownloadFailure()
		return err
	}
	m, err := ce.MD5()
	if err != nil {
		fmt.Printf("failed to get md5 from file: %s\n", err)
		ce.DownloadFailure()
		return err
	}

	md5url := url + ".md5"
	bs, err := gh.Get(md5url)
	if err != nil {
		fmt.Printf("failed to get md5 sum: %s\n", err)
		ce.DownloadFailure()
		return err
	}
	fs := strings.Fields(string(bs))
	m_download := fs[0]
	if m != m_download {
		fmt.Printf("Checksum file: %s\n", m)
		fmt.Printf("Checksum  url: %s\n", m_download)
		ce.DownloadFailure()
		return fmt.Errorf("checksum mismatch")
	}
	fmt.Printf("Downloaded (and checksum verified) %s\n", url)
	return nil

}
