package main

import (
	"fmt"
	pb "golang.conradwood.net/apis/commondeploy"
	"golang.conradwood.net/autodeployer/mkenv"
	"golang.conradwood.net/go-easyops/authremote"
)

func Mkenv() error {
	fmt.Printf("Setting up environment...\n")
	ctx := authremote.Context()
	mr := &pb.MkenvRequest{}
	_, err := mkenv.Setup(ctx, mr)
	return err
}
