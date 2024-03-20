package common

import (
	"golang.conradwood.net/apis/registry"
)

type AutodeployerGroup struct {
	deployers []*Deployer
}
type Deployer struct {
	adr *registry.ServiceAddress
}

func (d *Deployer) String() string {
	return d.adr.Host
}

func NewAutodeployerGroup(sas []*registry.ServiceAddress) *AutodeployerGroup {
	ag := &AutodeployerGroup{}
	for _, sa := range sas {
		ag.deployers = append(ag.deployers, &Deployer{adr: sa})
	}
	return ag
}
func (ag *AutodeployerGroup) FilterByMachine(machine string) *AutodeployerGroup {
	return ag
}
func (ag *AutodeployerGroup) Deployers() []*Deployer {
	return ag.deployers
}
