package mkenv

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/commondeploy"
	"golang.conradwood.net/autodeployer/fscache"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/utils"
	//	"os"
	"strings"
)

func Setup(ctx context.Context, req *pb.MkenvRequest) (*pb.MkenvResponse, error) {
	if !utils.FileExists(req.TargetDirectory) {
		return nil, fmt.Errorf("target directory \"%s\" does not exist", req.TargetDirectory)
	}
	fmt.Println(fscache.ListToString("/srv/autodeployer/fscache/cache_list_file"))
	oe := &oneenv{
		fscache: fscache.NewFSCache(1024*10, "/srv/autodeployer/fscache"),
		req:     req,
		ctx:     ctx,
		workdir: "/srv/autodeployer/mkenv/" + utils.RandomString(32),
	}
	utils.RecreateSafely(oe.workdir)
	oe.fscache.RegisterDeriveFunctionDir("untar", Derive_Untar)
	tardir, err := oe.CacheRootFS()
	if err != nil {
		return nil, err
	}
	oe.extracted_rootfs = tardir
	if oe.req.UseOverlayFS {
		err = oe.MountOverlayFS()
	} else {
		err = oe.CopyRootFS()
	}
	if err != nil {
		return nil, err
	}
	res := &pb.MkenvResponse{}
	return res, nil
}

// ensure that RootFS is cached on a local disk
func (oe *oneenv) CacheRootFS() (string, error) {
	cr := &pb.CacheRequest{URL: oe.req.RootFileSystemID}
	ce, err := oe.fscache.Cache(oe.ctx, cr)
	if err != nil {
		return "", err
	}

	_, err = oe.fscache.GetDerivedFile(ce, "rootfs.tar", "unbzip2")
	if err != nil {
		return "", err
	}

	tardir, err := oe.fscache.GetDerivedFileFromDerived(ce, "rootfs.tar", "rootfs", "untar")
	if err != nil {
		return "", err
	}
	/*
		err = Untar(tarname, oe.envdir+"/root")
		if err != nil {
			return err
		}
	*/
	fmt.Printf("root filesystem in \"%s\"\n", tardir)
	return tardir, nil
}

func (oe *oneenv) MountOverlayFS() error {
	lowerdir := oe.extracted_rootfs
	upperdir := oe.workdir + "/overlayfs/upper"
	utils.RecreateSafely(upperdir)
	workdir := oe.workdir + "/overlayfs/workdir"
	utils.RecreateSafely(workdir)
	target := oe.req.TargetDirectory
	fmt.Printf("Mounting overlayfs with lower=\"%s\" , upper=\"%s\" in %s\n", lowerdir, upperdir, target)
	l := linux.New()
	com := []string{
		"mount",
		"-t", "overlay",
		"overlay",
		"-o", fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerdir, upperdir, workdir),
		target,
	}
	fmt.Printf("Mounting: %s\n", strings.Join(com, " "))
	b, err := l.SafelyExecute(com, nil)
	if err != nil {
		fmt.Printf("Failed to mount:%s\n%s\n", err, b)
		return err
	}
	return fmt.Errorf("cannot do overlayfs yet")
}
func (oe *oneenv) CopyRootFS() error {
	return fmt.Errorf("cannot copy rootfs yet")
}
