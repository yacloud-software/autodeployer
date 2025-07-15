package main

import (
	"fmt"
	"os"
	"path/filepath"

	ad "golang.conradwood.net/apis/autodeployer"
	dc "golang.conradwood.net/deploymonkey/common"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/utils"
)

func DeployLocal() {
	conn, err := client.ConnectWithIP("localhost:4000")
	utils.Bail("failed to connect to local autodeployer", err)
	adc := ad.NewAutoDeployerClient(conn)
	rootdir := FindGitRoot()
	fmt.Printf("Git Root: %s\n", rootdir)
	fd, err := dc.ParseFile(rootdir+"/deployment/deploy.yaml", *repository)
	utils.Bail("failed to parse file", err)
	fmt.Printf("Namespace: %s\n", fd.Namespace)
	for _, group := range fd.Groups {
		for _, app := range group.Applications {
			// if !matches(group,app)  ->continue()
			deployRequest := dc.CreateDeployRequest(group, app)
			deployRequest.DeploymentID = "testing"
			deployRequest.BuildID = 123456
			fmt.Printf("DeployRequest:\n%#v\n", deployRequest)
			ctx := Context()
			adc.Deploy(ctx, deployRequest)
		}
	}
}

func FindGitRoot() string {
	path, err := os.Getwd()
	utils.Bail("failed to get current dir", err)
	path, err = filepath.Abs(path)
	utils.Bail("no git root", err)
	for {
		if utils.FileExists(path + "/deployment") {
			break
		}
		path = filepath.Dir(path)
	}
	return path
}
