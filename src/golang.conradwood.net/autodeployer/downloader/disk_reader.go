package downloader

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

var (
	disk_reader_counter = 0
	nextnumlock         sync.Mutex
)

type disk_reader struct {
	num               int
	entry             *diskCacheEntry
	fd                *os.File
	read_from_mem     bool // true if this one reads from memory
	next_byte_to_send int
	failure           error
}

func (d *disk_reader) Read(buf []byte) (int, error) {
	if d.failure != nil {
		return 0, d.failure
	}
	d.entry.lastRead = time.Now()
	if !d.read_from_mem {
		if d.fd == nil {
			fd, err := os.Open(d.entry.filename)
			if err != nil {
				return 0, err
			}
			d.fd = fd
		}
		return d.fd.Read(buf)
	}

	// this one is a bit complicated, we need to keep reading as data comes available in the cache
	n := len(buf)
	max := len(d.entry.partial_data)
	if (d.next_byte_to_send + n) > max {
		n = max - d.next_byte_to_send
	}
	for i := 0; i < n; i++ {
		buf[i] = d.entry.partial_data[d.next_byte_to_send]
		d.next_byte_to_send++
	}
	if n == 0 {
		time.Sleep(2 * time.Second)
	}
	//	d.Printf("Sent %d bytes, position %d of %d (download complete:%v)\n", n, d.next_byte_to_send, max, d.entry.completely_downloaded)
	if (n == 0) && (d.next_byte_to_send == len(d.entry.partial_data)) && d.entry.completely_downloaded {
		return 0, io.EOF
	}
	if d.entry.tobedeleted || d.entry.abort_download {
		d.failure = fmt.Errorf("Failed to download")
		return 0, d.failure
	}
	return n, nil
}

func (d *disk_reader) Close() error {
	if d.fd != nil {
		d.fd.Close()
		d.fd = nil
	}
	d.entry.removeReader(d)
	return nil
}
func (d *disk_reader) Printf(txt string, args ...interface{}) {
	s := fmt.Sprintf(txt, args...)
	fmt.Printf("[%d] %s", d.num, s)
}

func nextDiskReaderNum() int {
	nextnumlock.Lock()
	disk_reader_counter++
	a := disk_reader_counter
	nextnumlock.Unlock()
	return a
}
