package targets

import (
	"golang.conradwood.net/deployminator/db"
)

type TargetList struct {
	parent  *TargetList
	targets []*Target
}

func (tl *TargetList) FilterByMachineGroup(machinegroup string) *TargetList {
	res := &TargetList{parent: tl}
	for _, t := range tl.targets {
		if !t.IsInMachineGroup(machinegroup) {
			continue
		}
		res.targets = append(res.targets, t)
	}
	return res
}

// returns a target with no deployments or zero
func (tl *TargetList) TargetWithZeroDeployments() *Target {
	for _, t := range tl.targets {
		if t.AppsCount() == 0 {
			return t
		}
	}
	return nil
}

/* returns the target(s) which have the least amount of instances of this app.
for example:
Target 1: 3 instances of app
Target 2: 2 instances of app
Target 3: 2 instances of app
returns target 2 & 3

if Target 4: 1 instances of app, it would only return Target 4
*/
func (tl *TargetList) TargetsWithLeastInstancesOfApp(req *db.FullDD) *TargetList {
	minCount := -1
	for _, t := range tl.targets {
		c := t.CountInstancesOfApp(req)
		if minCount == -1 || c < minCount {
			minCount = c
		}
	}
	// now find all targets with EXACTLY minCount
	res := &TargetList{parent: tl}
	for _, t := range tl.targets {
		c := t.CountInstancesOfApp(req)
		if c == minCount {
			res.targets = append(res.targets, t)
		}
	}
	return res
}
func (tl *TargetList) TargetsWithLeastInstances() *TargetList {
	minCount := -1
	for _, t := range tl.targets {
		c := t.AppsCount()
		if minCount == -1 || c < minCount {
			minCount = c
		}
	}
	// now find all targets with EXACTLY minCount
	res := &TargetList{parent: tl}
	for _, t := range tl.targets {
		c := t.AppsCount()
		if c == minCount {
			res.targets = append(res.targets, t)
		}
	}
	return res
}

func (tl *TargetList) Count() int {
	return len(tl.targets)
}
func (tl *TargetList) Targets() []*Target {
	return tl.targets
}
