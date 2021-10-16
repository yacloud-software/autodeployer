package main

import (
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/autodeployer"
	"sync"
)

var (
	webdir  = flag.String("webdir", "/var/www/static", "Web directory to deploy webpackages into")
	webLock sync.Mutex
)

// this deploys a webpackage
// it runs within the autodeployer-server process space
// this *should* be async?
func DeployWebPackage(cr *pb.DeployRequest) error {
	return fmt.Errorf("deploying webpackages not (no longer) supported")
}
