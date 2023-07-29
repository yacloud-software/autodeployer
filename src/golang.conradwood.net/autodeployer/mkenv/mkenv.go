package mkenv

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/commondeploy"
	"golang.conradwood.net/autodeployer/fscache"
	"golang.conradwood.net/go-easyops/utils"
	"os"
)

func Setup(ctx context.Context, req *pb.MkenvRequest) (*pb.MkenvResponse, error) {
	fmt.Println(fscache.ListToString("/srv/autodeployer/fscache/cache_list_file"))
	oe := &oneenv{
		fscache: fscache.NewFSCache(1024*10, "/srv/autodeployer/fscache"),
		req:     req,
		ctx:     ctx,
		envdir:  "/srv/autodeployer/rootfs/" + utils.RandomString(32),
	}
	oe.fscache.RegisterDeriveFunctionDir("untar", Derive_Untar)
	os.MkdirAll(oe.envdir, 0777)
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

	_, err = oe.fscache.GetDerivedFile(ce, "rootfs.tar", "unbzip2")
	if err != nil {
		return err
	}

	tardir, err := oe.fscache.GetDerivedFileFromDerived(ce, "rootfs.tar", "rootfs", "untar")
	if err != nil {
		return err
	}
	/*
		err = Untar(tarname, oe.envdir+"/root")
		if err != nil {
			return err
		}
	*/
	fmt.Printf("root filesystem in \"%s\"\n", tardir)
	return nil
}
