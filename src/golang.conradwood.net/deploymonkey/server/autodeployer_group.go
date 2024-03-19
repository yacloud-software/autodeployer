package main

import (
	"golang.conradwood.net/apis/registry"
)

type AutodeployerGroup struct {
	deployers []*deployer
}
type deployer struct {
	adr *registry.ServiceAddress
}

func (d *deployer) String() string {
	return d.adr.Host
}

func NewAutodeployerGroup(sas []*registry.ServiceAddress) *AutodeployerGroup {
	ag := &AutodeployerGroup{}
	for _, sa := range sas {
		ag.deployers = append(ag.deployers, &deployer{adr: sa})
	}
	return ag
}
func (ag *AutodeployerGroup) FilterByMachine(machine string) *AutodeployerGroup {
	return ag
}
func (ag *AutodeployerGroup) Deployers() []*deployer {
	return ag.deployers
}
