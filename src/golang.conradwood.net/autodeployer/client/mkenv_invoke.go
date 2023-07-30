package main

import (
	"fmt"
	pb "golang.conradwood.net/apis/commondeploy"
	"golang.conradwood.net/autodeployer/mkenv"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
)

func Mkenv() error {
	fmt.Printf("Setting up environment...\n")
	ctx := authremote.Context()
	mr := &pb.MkenvRequest{
		RootFileSystemID: "http://johnsmith/rootfs.tar.bz2",
		UseOverlayFS:     true,
		TargetDirectory:  "/srv/temp/rootfs/exe_dir",
	}
	utils.RecreateSafely(mr.TargetDirectory)
	m := mkenv.NewMkenv("/srv/temp/mkenv", true)
	err := m.UnmountAll()
	if err != nil {
		fmt.Printf("WARNING: failed to unmount (%s)\n", err)
	}
	_, err = m.Setup(ctx, mr)
	return err
}
