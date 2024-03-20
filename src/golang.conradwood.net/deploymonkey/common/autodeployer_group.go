package common

import (
	"golang.conradwood.net/apis/registry"
	"sync"
)

var (
	deployers = make(map[string]*Deployer)
	deplock   sync.Mutex
)

type AutodeployerGroup struct {
	deployers []*Deployer
}
type Deployer struct {
	adr *registry.ServiceAddress
}

func (d *Deployer) Host() string {
	return d.adr.Host
}
func (d *Deployer) String() string {
	return d.adr.Host
}

func NewAutodeployerGroup(sas []*registry.ServiceAddress) *AutodeployerGroup {
	deplock.Lock()
	defer deplock.Unlock()
	ag := &AutodeployerGroup{}
	for _, sa := range sas {
		d, fd := deployers[sa.Host]
		if !fd {
			d = &Deployer{adr: sa}
			deployers[sa.Host] = d
		}
		d.adr = sa
		ag.deployers = append(ag.deployers, d)
	}
	return ag
}
func (ag *AutodeployerGroup) FilterByMachine(machine string) *AutodeployerGroup {
	return ag
}
func (ag *AutodeployerGroup) Deployers() []*Deployer {
	return ag.deployers
}
