package changes

import (
	"fmt"
	"golang.conradwood.net/deployminator/db"
	"golang.conradwood.net/deployminator/targets"
	"sync"
)

var (
	goals = make(map[uint64]*Goal)
	glock sync.Mutex
)

// specific deployments for a given deployment descriptor
type Goal struct {
	deployments []*DeployGoal
	Broken      bool // true if this goal cannot be achieved (and is probably incomplete)
}
type DeployGoal struct {
	target targets.Target
	fulldd *db.FullDD
}

func define_goal(req *db.FullDD) *Goal {
	glock.Lock()
	defer glock.Unlock()
	g := goals[req.DeploymentDescriptor.ID]

	if g == nil {
		g = &Goal{}
		goals[req.DeploymentDescriptor.ID] = g
	}

	g.Broken = false
	tl := targets.GetTargets()
	// calculate what we want...
	for _, id := range req.Instances {
		tlm := tl.FilterByMachineGroup(id.Instance.MachineGroup)
		if id.Instance.InstanceCountIsPerMachine {
			panic("InstanceCount not yet implemented")
		}
		if tlm.Count() == 0 {
			fmt.Printf("We got no targets for machinegroup \"%s\"\n", id.Instance.MachineGroup)
			g.Broken = true
			continue
		}
		t := tlm.TargetWithZeroDeployments()
		if t != nil {
			t.AddPendingApp(0, req)
			continue
		}
		tlm = tlm.TargetsWithLeastInstancesOfApp(req)
		tlm = tlm.TargetsWithLeastInstances()
		if tlm.Count() == 0 {
			fmt.Printf("no instance with least stuff - bug?\n")
			g.Broken = true
			return g
		}
		t = tlm.Targets()[0]
		t.AddPendingApp(0, req)
	}
	return g
}
