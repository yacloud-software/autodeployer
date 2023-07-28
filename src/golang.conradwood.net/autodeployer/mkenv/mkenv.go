package mkenv

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/commondeploy"
	"golang.conradwood.net/autodeployer/fscache"
	"io"
)

func Setup(ctx context.Context, req *pb.MkenvRequest) (*pb.MkenvResponse, error) {
	fmt.Println(fscache.ListToString("/srv/autodeployer/fscache/cache_list_file"))
	oe := &oneenv{
		fscache: fscache.NewFSCache(1024*10, "/srv/autodeployer/fscache"),
		req:     req,
		ctx:     ctx}
	err := oe.CacheRootFS()
	if err != nil {
		return nil, err
	}
	res := &pb.MkenvResponse{}
	return res, nil
}

// ensure that RootFS is cached on a local disk
func (oe *oneenv) CacheRootFS() error {
	cr := &pb.CacheRequest{URL: "http://johnsmith/rootfs.tar.bz2"}
	ce, err := oe.fscache.Cache(oe.ctx, cr)
	if err != nil {
		return err
	}

	unbzip2_function := func(r io.Reader, w io.Writer) error {
		_, err := io.Copy(w, r)
		return err
	}
	oe.fscache.RegisterDeriveFunction("unbzip2", unbzip2_function)

	_, err = oe.fscache.GetDerivedFile(ce, "rootfs.tar", "unbzip2")
	if err != nil {
		return err
	}
	return nil

}
