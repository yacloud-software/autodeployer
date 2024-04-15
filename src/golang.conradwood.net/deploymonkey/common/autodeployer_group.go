package common

import (
	"context"
	ad "golang.conradwood.net/apis/autodeployer"
	"golang.conradwood.net/apis/registry"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/client"
	"google.golang.org/grpc"
	"strings"
	"sync"
	"time"
)

var (
	deployers  = make(map[string]*Deployer)
	deplock    sync.Mutex
	kick_query = make(chan string)
)

func init() {
	go query_autodeployers_loop()
}

type AutodeployerGroup struct {
	deployers []*Deployer
}
type Deployer struct {
	is_online                  bool
	failed_deployments         int
	last_successful_deployment time.Time
	first_failed_deployment    time.Time
	adr                        *registry.ServiceAddress
	con                        *grpc.ClientConn
	lock                       sync.Mutex
	machine_info               *ad.MachineInfoResponse
	lastInfoResponse           *ad.InfoResponse
	info_response_retrieved_at time.Time
	currently_refreshing       bool
}

func DeployerByIP(ip string) *Deployer {
	deplock.Lock()
	defer deplock.Unlock()
	for k, d := range deployers {
		if k == ip {
			return d
		}
	}
	return nil
}

// log a failed deployment. Eventually a deployer will be excluded if it continues to fail
func (d *Deployer) DeploymentFailed() {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.failed_deployments == 0 {
		d.first_failed_deployment = time.Now()
	}
	d.failed_deployments++
}

// reset the deployer failure counter
func (d *Deployer) DeploymentSucceeded() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.failed_deployments = 0
	d.last_successful_deployment = time.Now()
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
func (d *Deployer) GetDeployments() *DeploymentGroup {
	res := NewDeploymentGroup()
	res.AddDeployment(d, d.lastInfoResponse, d.info_response_retrieved_at)
	return res
}
func (d *Deployer) AppByID(id string) *ad.DeployedApp {
	lir := d.lastInfoResponse
	if lir == nil {
		return nil
	}

	for _, app := range lir.Apps {
		if app.ID == id {
			return app
		}
	}
	return nil
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
	ir, err := d.GetClient().GetDeployments(ctx, CreateInfoRequest())
	if err != nil {
		d.con.Close()
		d.con = nil
		return err
	}
	d.lastInfoResponse = ir
	d.info_response_retrieved_at = time.Now()

	return nil
}

// re-requery the deployer about its deployments
func (d *Deployer) refresh_deployments(ctx context.Context) {
	if d.currently_refreshing {
		// despite race-condition should keep refresh thread count lowish
		return
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	d.currently_refreshing = true
	ir, err := d.GetClient().GetDeployments(ctx, CreateInfoRequest())
	if err != nil {
		d.currently_refreshing = false
		// failed to query
		return
	}
	d.lastInfoResponse = ir
	d.info_response_retrieved_at = time.Now()
	d.currently_refreshing = false
	return

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

func query_autodeployers_loop() {
	for {
		s := ""
		select {
		case s = <-kick_query:
			//
		case <-time.After(time.Duration(15) * time.Second):
			//
		}
		if s != "" {
			deplock.Lock()
			depl := deployers[s]
			deplock.Unlock()
			if depl == nil {
				continue
			}
			ctx := authremote.Context()
			go depl.refresh_deployments(ctx)
			continue
		}

		var depls []*Deployer
		deplock.Lock()
		for _, d := range deployers {
			depls = append(depls, d)
		}
		deplock.Unlock()
		wg := &sync.WaitGroup{}
		ctx := authremote.Context()
		for _, depl := range depls {
			wg.Add(1)
			go func(d *Deployer) {
				d.refresh_deployments(ctx)
				wg.Done()
			}(depl)
		}
		wg.Wait()
	}
}
