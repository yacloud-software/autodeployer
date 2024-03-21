package common

import (
	"golang.conradwood.net/apis/registry"
	"golang.conradwood.net/go-easyops/client"
	"google.golang.org/grpc"
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
	adr  *registry.ServiceAddress
	con  *grpc.ClientConn
	lock sync.Mutex
}

func (d *Deployer) Host() string {
	return d.adr.Host
}
func (d *Deployer) String() string {
	return d.adr.Host
}
func (d *Deployer) GetConnection() (*grpc.ClientConn, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.con != nil {
		return d.con, nil
	}
	con, err := client.ConnectWithIP(d.adr.Host + ":4000")
	if err != nil {
		return nil, err
	}
	d.con = con
	return d.con, nil
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
