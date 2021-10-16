package targets

import (
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	"golang.conradwood.net/deployminator/db"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/client"
	"google.golang.org/grpc"
	"strings"
	"sync"
)

type PendingApp struct {
	Key uint64
	dd  *db.FullDD
}

// keep track of which autodeployers we have
type Target struct {
	lock          sync.Mutex
	address       string
	machinegroups []string
	con           *grpc.ClientConn
	adclient      ad.AutoDeployerClient
	apps          []*ad.DeployedApp // apps running, actually deployed
	pendingApps   []*PendingApp     // apps pending to be started
	//PendingApps []uint64 // apps pending to be started
}

func (t *Target) Host() string {
	return t.address
}
func (t *Target) AddPendingApp(key uint64, dd *db.FullDD) {
	t.pendingApps = append(t.pendingApps, &PendingApp{Key: key, dd: dd})
}
func (t *Target) GetPendingApps() []*PendingApp {
	return t.pendingApps
}
func (t *Target) ClearPendingApps() {
	t.pendingApps = t.pendingApps[:0]
}
func (t *Target) String() string {
	return fmt.Sprintf("%s(%s)", t.address, strings.Join(t.machinegroups, ","))
}

func (t *Target) RemovePendingApp(key uint64) {
	if t == nil {
		return
	}
	var res []*PendingApp
	for _, pa := range t.pendingApps {
		if pa.Key == key {
			continue
		}
		res = append(res, pa)
	}
	t.pendingApps = res
}

func (t *Target) init() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.con == nil {
		con, err := client.ConnectWithIP(t.address)
		if err != nil {
			return err
		}
		t.con = con
		t.adclient = ad.NewAutoDeployerClient(t.con)
	}
	return nil
}
func (t *Target) Autodeployer() ad.AutoDeployerClient {
	return t.adclient
}
func (t *Target) Scan() error {
	err := t.init()
	if err != nil {
		return err
	}
	ctx := authremote.Context()
	mr, err := t.adclient.GetMachineInfo(ctx, &ad.MachineInfoRequest{})
	if err != nil {
		return err
	}
	t.machinegroups = mr.MachineGroup
	depls, err := t.adclient.GetDeployments(ctx, &ad.InfoRequest{})
	if err != nil {
		return err
	}
	t.apps = depls.Apps
	if *debug {
		fmt.Printf("   Machinegroups: %v\n", t.machinegroups)
		for _, d := range t.apps {
			fmt.Printf("   %s/%s\n", d.Deployment.Binary, d.Deployment.Status)
		}
	}
	return nil
}

func (t *Target) IsInMachineGroup(machinegroup string) bool {
	for _, mg := range t.machinegroups {
		if mg == machinegroup {
			return true
		}
	}
	return false
}

func (t *Target) AppsCount() int {
	return len(t.apps) + len(t.pendingApps)
}
func (t *Target) CountInstancesOfApp(req *db.FullDD) int {
	res := 0
	for _, app := range t.pendingApps {
		if app.dd.DeploymentDescriptor.ID == req.DeploymentDescriptor.ID {
			res++
		}
	}
	for _, app := range t.apps {
		if matchesDD(req, app) {
			res++
		}
	}
	return res
}

func matchesDD(dd *db.FullDD, app *ad.DeployedApp) bool {
	dr := app.DeployRequest
	a := dd.DeploymentDescriptor.Application
	if dr.Binary != a.Binary {
		return false
	}
	if dr.BuildID != dd.DeploymentDescriptor.BuildNumber {
		return false
	}
	if dr.RepositoryID != a.RepositoryID {
		return false
	}
	return true

}
