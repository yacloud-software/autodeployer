package fscache

import (
	pb "golang.conradwood.net/apis/commondeploy"
	"golang.conradwood.net/go-easyops/utils"
	"os"
	"path/filepath"
	"strings"
)

type cache_list struct {
	filename string
	entries  []*pb.CacheEntry
}

// reads (or creates) file
func read_cache_list_file(filename string) (*cache_list, error) {
	var err error
	var cache_list_contents []byte
	if !utils.FileExists(filename) {
		cache_list_contents = make([]byte, 0)
		b := filepath.Dir(filename)
		os.MkdirAll(b, 0777)
		err = utils.WriteFile(filename, cache_list_contents)
	} else {
		cache_list_contents, err = utils.ReadFile(filename)
	}
	if err != nil {
		return nil, err
	}
	res := &cache_list{
		filename: filename,
	}
	for _, line := range strings.Split(string(cache_list_contents), "\n") {
		if len(line) < 3 {
			continue
		}
		ce := &pb.CacheEntry{}
		err = utils.Unmarshal(line, ce)
		if err != nil {
			return nil, err
		}
		res.entries = append(res.entries, ce)
	}
	return res, nil
}

func (c *cache_list) add(ce *pb.CacheEntry) {
	c.entries = append(c.entries, ce)
}
func (c *cache_list) write() error {
	var content string

	for _, ce := range c.entries {
		s, err := utils.Marshal(ce)
		if err != nil {
			return err
		}
		content = content + s + "\n"
	}
	err := utils.WriteFile(c.filename, []byte(content))
	return err
}
