package main

import (
	"fmt"
	pb "golang.conradwood.net/apis/autodeployer"
	"sync"
)

var (
	staticLock sync.Mutex
)

// this deploys a webpackage
// it runs within the autodeployer-server process space
// this *should* be async?
func DeployStaticFiles(cr *pb.DeployRequest) error {
	return fmt.Errorf("deploying webpackages not (no longer) supported")
}
