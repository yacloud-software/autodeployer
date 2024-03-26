package common

import (
	ad "golang.conradwood.net/apis/autodeployer"
	"golang.conradwood.net/apis/registry"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/client"
	"google.golang.org/grpc"
	"strings"
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
	adr          *registry.ServiceAddress
	con          *grpc.ClientConn
	lock         sync.Mutex
	machine_info *ad.MachineInfoResponse
}

func (d *Deployer) Host() string {
	return d.adr.Host
}
func (d *Deployer) String() string {
	return d.adr.Host
}
func (d *Deployer) GetClient() ad.AutoDeployerClient {
	return ad.NewAutoDeployerClient(d.con)
}
func (d *Deployer) GetMachineGroups() []string {
	return d.machine_info.MachineGroup
}
func (d *Deployer) ServesMachineGroup(machinegroupname string) bool {
	if machinegroupname == "" {
		//nothing -> match anything
		return true
	}
	if strings.Contains(machinegroupname, "*") {
		//asterisk -> match anything
		return true
	}
	for _, m := range d.GetMachineGroups() {
		if m == machinegroupname {
			return true
		}
	}
	return false
}
func (d *Deployer) init() error {
	if d.con == nil {
		con, err := client.ConnectWithIP(d.adr.Host + ":4000")
		if err != nil {
			return err
		}
		d.con = con
	}
	ctx := authremote.Context()
	mir, err := d.GetClient().GetMachineInfo(ctx, &ad.MachineInfoRequest{})
	if err != nil {
		d.con.Close()
		d.con = nil
		return err
	}
	d.machine_info = mir
	return nil
}
func NewAutodeployerGroup(sas []*registry.ServiceAddress) (*AutodeployerGroup, error) {
	deplock.Lock()
	defer deplock.Unlock()
	ag := &AutodeployerGroup{}
	for _, sa := range sas {
		d, fd := deployers[sa.Host]
		if !fd {
			d = &Deployer{adr: sa}
			deployers[sa.Host] = d
		}
		err := d.init()
		if err != nil {
			return nil, err
		}
		d.adr = sa
		ag.deployers = append(ag.deployers, d)
	}
	return ag, nil
}
func (ag *AutodeployerGroup) clone() *AutodeployerGroup {
	res := &AutodeployerGroup{
		deployers: ag.deployers,
	}
	return res
}
func (ag *AutodeployerGroup) FilterByMachine(machine string) *AutodeployerGroup {
	var depl []*Deployer
	for _, d := range ag.deployers {
		if !d.ServesMachineGroup(machine) {
			continue
		}
		depl = append(depl, d)
	}
	nag := ag.clone()
	nag.deployers = depl
	return nag
}
func (ag *AutodeployerGroup) Deployers() []*Deployer {
	return ag.deployers
}
