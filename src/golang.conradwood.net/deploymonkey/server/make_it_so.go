package main

import (
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	//	"golang.conradwood.net/apis/registry"
	//	"golang.conradwood.net/deploymonkey/common"
	"golang.conradwood.net/deploymonkey/deployplacements"
	"golang.conradwood.net/deploymonkey/deployq"
	"sync"
)

const (
	CACHE_WORKERS = 3
)

var (
	cache_worker_chan = make(chan *cachereq)
)

type cachereq struct {
	url string
	wg  *sync.WaitGroup
	err error
}

func (cr *cachereq) String() string {
	return fmt.Sprintf("[cachereq for %s]", cr.url)
}

func init() {
	for i := 0; i < CACHE_WORKERS; i++ {
		go cache_worker_loop()
	}
}

func makeitso_new(group *DBGroup, apps []*pb.ApplicationDefinition) error {
	fmt.Printf("[newstyle] deploying %d apps in new_style\n", len(apps))
	sas, err := GetDeployers()
	if err != nil {
		return err
	}

	// step #1 - build up a list what we want to deploy on which autodeployer
	var deployments []*deployplacements.DeployRequest
	for _, app := range apps {
		drs, err := deployplacements.Create_requests_for_app(group, app, sas)
		if err != nil {
			return err
		}
		deployments = append(deployments, drs...)
	}
	fmt.Printf("[newstyle] %d deployments:\n", len(deployments))
	for _, d := range deployments {
		fmt.Printf("[newstyle] Deploy: %s\n", d.String())
	}

	// now we got a bunch of deployment requests, handle them
	deployq.Add(deployments)

	return nil
}
func cache_worker_loop() {
	for {
		cr := <-cache_worker_chan
		fmt.Printf("Caching %s\n", cr.String())
	}
}
