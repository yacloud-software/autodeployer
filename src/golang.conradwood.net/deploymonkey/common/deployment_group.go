package common

import (
	ad "golang.conradwood.net/apis/autodeployer"
	"sync"
	"time"
)

// a DeploymentGroup has information about actual deployments on autodeployers
type DeploymentGroup struct {
	deployments map[string]*ad_deployer_info // ip->link
	lock        sync.Mutex
}
type ad_deployer_info struct {
	deployer  *Deployer
	retrieved time.Time
	response  *ad.InfoResponse
}

func NewDeploymentGroup() *DeploymentGroup {
	return &DeploymentGroup{deployments: make(map[string]*ad_deployer_info)}
}
func (d *DeploymentGroup) AddDeployment(depl *Deployer, ad *ad.InfoResponse, retrievalDate time.Time) {
	d.lock.Lock()
	defer d.lock.Unlock()
	adi := d.deployments[depl.Host()]
	if adi == nil {
		adi = &ad_deployer_info{}
		d.deployments[depl.Host()] = adi
	}
	adi.deployer = depl
	adi.response = ad
	adi.retrieved = retrievalDate
}
