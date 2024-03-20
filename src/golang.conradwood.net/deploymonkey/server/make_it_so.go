package main

import (
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/apis/registry"
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

type deploy_request struct {
	appdef *pb.ApplicationDefinition
	sa     *deployer
}

func (dr *deploy_request) String() string {
	return fmt.Sprintf("%s on %s", dr.appdef.Binary, dr.sa.String())
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
	var deployments []*deploy_request
	for _, app := range apps {
		drs, err := create_requests_for_app(group, app, sas)
		if err != nil {
			return err
		}
		deployments = append(deployments, drs...)
	}
	fmt.Printf("[newstyle] %d deployments:\n", len(deployments))
	for _, d := range deployments {
		fmt.Printf("Deploy: %s\n", d.String())
	}
	return nil
}
func create_requests_for_app(group *DBGroup, app *pb.ApplicationDefinition, sas []*registry.ServiceAddress) ([]*deploy_request, error) {
	var res []*deploy_request
	ag := NewAutodeployerGroup(sas)
	if app.InstancesMeansPerAutodeployer {
		for _, s := range ag.FilterByMachine(app.Machines).Deployers() {
			for i := 0; i < int(app.Instances); i++ {
				dr := &deploy_request{
					appdef: app,
					sa:     s,
				}
				res = append(res, dr)
			}
		}
		return res, nil
	}

	for len(res) < int(app.Instances) {
		for _, s := range ag.FilterByMachine(app.Machines).Deployers() {
			dr := &deploy_request{
				appdef: app,
				sa:     s,
			}
			res = append(res, dr)
			if len(res) >= int(app.Instances) {
				break
			}

		}
	}
	return res, nil

}
func cache_worker_loop() {
	for {
		cr := <-cache_worker_chan
		fmt.Printf("Caching %s\n", cr.String())
	}
}
