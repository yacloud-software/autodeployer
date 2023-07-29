package fscache

import (
	"fmt"
	pb "golang.conradwood.net/apis/commondeploy"
)

func ListToString(filename string) string {
	cl, err := read_cache_list_file(filename)
	if err != nil {
		return fmt.Sprintf("%s", err)
	}
	s := "Cachelist " + filename + "\n"
	for _, ce := range cl.entries {
		s = s + render_cache_entry(" ", ce)
	}
	return s
}
func render_cache_entry(prefix string, ce *pb.CacheEntry) string {
	s := prefix + "CacheEntry for " + ce.CachedURL + "\n"
	s = s + prefix + fmt.Sprintf("  Downloading: %v\n", ce.Downloading)
	s = s + prefix + fmt.Sprintf("  Downloaded : %v\n", ce.Downloaded)
	s = s + prefix + fmt.Sprintf("  CacheDir   : %s\n", ce.CacheDir)
	for _, dce := range ce.DerivedEntries {
		s = s + render_derived_entry(prefix+"   ", dce)
	}
	return s
}
func render_derived_entry(prefix string, dce *pb.DerivedCacheEntry) string {
	s := prefix + "DerivedCacheEntry \"" + dce.FileID + "\"\n"
	s = s + prefix + fmt.Sprintf("  FileRef   : %s\n", dce.FileRef)
	s = s + prefix + fmt.Sprintf("  Function  : %s\n", dce.Function)
	s = s + prefix + fmt.Sprintf("  Deriving  : %v\n", dce.Deriving)
	s = s + prefix + fmt.Sprintf("  Completed : %v\n", dce.Completed)
	return s
}
