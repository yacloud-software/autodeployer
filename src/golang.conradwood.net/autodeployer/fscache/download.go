package fscache

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	pb "golang.conradwood.net/apis/commondeploy"
	"golang.conradwood.net/go-easyops/http"
	"golang.conradwood.net/go-easyops/utils"
	"hash"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	BUFSIZE = 32768
)

func (f *fscache) download(ce *pb.CacheEntry) error {
	err := f.incActiveDownload()
	if err != nil {
		return err
	}
	defer f.decActiveDownload()

	fname := f.get_cache_dir(ce) + "/orig_file"
	// get a url
	f.debugf("Downloading %s to %s", ce.CachedURL, fname)
	if strings.HasPrefix(ce.CachedURL, "http") {
		return f.download_http(ce, fname)
	}
	return fmt.Errorf("unhandled url (%s)", ce.CachedURL)
}
func (f *fscache) download_http(ce *pb.CacheEntry, filename string) error {
	h := http.NewDirectClient()
	hr := h.Get(ce.CachedURL + ".md5")
	md5sum := ""
	err := hr.Error()
	if err == nil && hr.HTTPCode() >= 200 && hr.HTTPCode() < 300 {
		md5sum = cleanDownloadedMD5Sum(string(hr.Body()))
	}
	f.debugf("md5sum for %s: \"%s\"", ce.CachedURL, md5sum)

	h = http.NewDirectClient()
	hr = h.Head(ce.CachedURL)
	err = hr.Error()
	filesize := uint64(0)

	if err == nil && hr.HTTPCode() >= 200 && hr.HTTPCode() < 300 {
		hd := hr.Header("content-length")
		if hd != "" {
			f.debugf("Filesize: \"%s\"", hd)
			filesize, _ = strconv.ParseUint(hd, 10, 64)
		}
	}

	h = http.NewDirectClient()
	hr = h.GetStream(ce.CachedURL)
	err = hr.Error()
	if err != nil {
		return err
	}
	if hr.HTTPCode() < 200 || hr.HTTPCode() >= 300 {
		return fmt.Errorf("failed to download (code %d)", hr.HTTPCode())
	}
	started := time.Now()
	size, chk, err := f.streamBodyToFile(ce, filesize, hr.BodyReader(), filename)
	if err != nil {
		return err
	}
	if md5sum != "" && chk.String() != md5sum {
		return fmt.Errorf("checksum failed to verify (reported: \"%s\", downloaded: \"%s\")", md5sum, chk.String())
	}
	dur := time.Since(started)
	f.debugf("Downloaded %s (%d bytes in %0.1fs)", ce.CachedURL, size, dur.Seconds())
	return nil
}
func (f *fscache) incActiveDownload() error {
	f.d_lock.Lock()
	defer f.d_lock.Unlock()
	if f.active_downloads >= f.maxdownloads {
		return fmt.Errorf("too many (%d) downloads already in progress. (max: %d)", f.active_downloads, f.maxdownloads)
	}
	f.active_downloads++
	return nil
}
func (f *fscache) decActiveDownload() {
	f.d_lock.Lock()
	defer f.d_lock.Unlock()
	if f.active_downloads == 0 {
		panic("BUG in fscache. activedownloads is 0 and attempted to decrement")
	}
	f.active_downloads--
}

func (f *fscache) streamBodyToFile(ce *pb.CacheEntry, filesize uint64, r io.Reader, filename string) (int, *checksum, error) {
	fw, err := os.Create(filename)
	if err != nil {
		return 0, nil, err
	}
	size, c, err := f.streamBodyToWriter(ce, filesize, r, fw)
	fw.Close()
	return size, c, err

}

func (f *fscache) streamBodyToWriter(ce *pb.CacheEntry, filesize uint64, r io.Reader, w io.Writer) (int, *checksum, error) {
	p := &utils.ProgressReporter{}
	if filesize != 0 {
		p.SetTotal(filesize)
	}
	buf := make([]byte, BUFSIZE)
	size := 0
	m := md5.New()
	for {
		n, r_err := r.Read(buf)
		size = size + n
		p.Add(uint64(n))
		p.Print()
		rb := buf[:n]
		m.Write(rb)
		n, err := w.Write(rb)
		if n != len(rb) {
			return 0, nil, fmt.Errorf("short write (%d vs %d)", n, len(rb))
		}
		if err != nil {
			return 0, nil, err
		}
		if r_err == io.EOF {
			break
		}
		if r_err != nil {
			return 0, nil, err
		}
	}
	hash := m.Sum(nil)
	s := hex.EncodeToString(hash)
	c := &checksum{h: m, hs: s}
	return size, c, nil
}

type checksum struct {
	h  hash.Hash
	hs string
}

func (c *checksum) String() string {
	return c.hs

}

func cleanDownloadedMD5Sum(md5sum string) string {
	res := strings.Trim(md5sum, "\n")
	idx := strings.Index(res, " ")
	if idx != -1 {
		res = res[:idx]
	}
	return res
}
