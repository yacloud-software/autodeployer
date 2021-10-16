package changes

import (
	"fmt"
	pb "golang.conradwood.net/apis/deployminator"
	"golang.conradwood.net/deployminator/targets"
	"strings"
	//	"golang.conradwood.net/go-easyops/utils"
	"flag"
	"golang.conradwood.net/deployminator/db"
)

type ACTION int

const (
	ACTION_START_ON_SPECIFIC_MACHINE ACTION = iota
	ACTION_START_ANY_IN_MACHINE_GROUP
	ACTION_STOP_ON_SPECIFIC_MACHINE
	ACTION_STOP_ON_ANY_IN_MACHINE_GROUP
)

var (
	debug = flag.Bool("debug_changes", false, "debug changes")
)

type Change struct {
	action ACTION
}

func (c *Change) String() string {
	return fmt.Sprintf("%v", c.action)
}

type Resolver struct {
	req *db.FullDD
}

/*
this is subtly complex. this matches 'deployed instances' with 'configured instances'. There often is more than one way to match it.
for example an application might be configured like so:
2x Instances on machines 'worker_a"
AND 3x Instances on machines 'worker_a or worker_b".
It is important that the resolver is deterministic and that the changes it suggests result in a resolution!
*/
func find_need_starting(req *db.FullDD) ([]*Change, error) {
	resolver := &Resolver{req: req}
	return resolver.findChange()
}
func (r *Resolver) findChange() ([]*Change, error) {
	r.Debugf("Scanning descriptor #%d (%s)\n", r.req.DeploymentDescriptor.ID, r.req.DeploymentDescriptor.Application.Binary)

	if len(r.req.Instances) == 0 {
		return nil, nil
	}
	//	g := define_goal(req)
	ml := targets.GetMatchList()
	ml = ml.FilterByAppBinary(r.req.DeploymentDescriptor.Application.Binary)
	ml = ml.FilterByAppBuild(r.req.DeploymentDescriptor.BuildNumber)
	ml = ml.FilterByAppRepo(r.req.DeploymentDescriptor.Application.RepositoryID)
	remaining_apps := ml.Apps() // list of matching apps currently deployed
	r.Debugf("  found %d deployments for this app:\n", len(remaining_apps))
	for _, ma := range remaining_apps {
		app := ma.App()
		mgs := strings.Join(ma.HostMachineGroups(), " ")
		r.Debugf("  On %s (%s), Binary: %s, Build: %d, Repo:%d\n", ma.Host(), mgs, app.Deployment.Binary, app.Deployment.BuildID, app.Deployment.RepositoryID)
	}

	// we now check if we need to deploy more.
	// to do so, we go through instance definitions ("per machine" first) and remove instances from the list when "used" to satisfy an instancedef
	var res []*Change
	for _, i := range r.req.Instances {
		if i.Instance.InstanceCountIsPerMachine {
			panic("Per-Instance Count ist not fully implemented yet")
		}
	}
	// no per-machinegroup. so we can simply count instances per machinegroup:
	var ct []*instancematcher
	for _, r := range r.req.Instances {
		ct = append(ct, &instancematcher{instance: r.Instance})
	}

	r.Debugf("now resolving %d instance matchers\n", len(ct))
	// we use those which have a single machinegroup first
	for _, ma := range remaining_apps {
		if len(ma.MatchingMachineGroups()) == 0 {
			r.Debugf("Error: no matching groups (configured: \"%s\" vs autodeployer: \"%s\")\n", strings.Join(ma.ConfiguredMachineGroups(), ","), strings.Join(ma.TargetMachineGroups(), ","))
			continue
		}

		if len(ma.MatchingMachineGroups()) < 1 {
			continue
		}
		mag := ma.ConfiguredMachineGroups()[0]
		if mag == "" {
			r.Debugf("   invalid configured machinegroup: \"%s\"\n", ma.App().Deployment.AppReference.AppDef.Machines)
			continue
		}
		var im *instancematcher
		for _, c := range ct {
			if c.isUsed() {
				continue
			}
			if !c.MatchesMachine(mag) {
				r.Debugf("   not a match (\"%s\" vs  \"%s\" <%s>)\n", c.instance.MachineGroup, mag, ma.App().Deployment.AppReference.AppDef.Machines)
				continue
			}
			r.Debugf("  added to instance %d\n", c.instance.ID)
			c.add(ma)
			im = c
			break
		}
		if im == nil {
			r.Printf("no matching instance definition for instance (\"%s\", autodeployer:\"%s\")\n", mag, strings.Join(ma.TargetMachineGroups(), ","))
		}
	}

	// check if all instance matchers are full
	for _, im := range ct {
		r.Debugf("Instance %s has %d deployments, and wants %d\n", im.String(), len(im.matches), im.instance.Instances)
		if !im.isUsed() {
			submit_deploy(im.instance)
			if *debug {
				r.Printf("Instance %s is not satisfied (has %d deployments, but wants %d)\n", im.String(), len(im.matches), im.instance.Instances)
			}
		}
	}
	return res, nil
}

func need_start_satisfy(res []*Change, id *pb.InstanceDef, ml *targets.MatchList) {
	// take them away from machines
}

// this matches instances to matchapps
type instancematcher struct {
	instance *pb.InstanceDef
	matches  []*targets.MatchApp
}

func (i *instancematcher) MatchesMachine(machine string) bool {
	cfg := i.instance.MachineGroup
	if cfg == "" {
		cfg = "worker"
	}
	return machine == cfg
}

func (i *instancematcher) String() string {
	return fmt.Sprintf("InstanceID=%d on %s", i.instance.ID, i.instance.MachineGroup)
}

// how many instances are missing (odd behavior if used with perinstancecount
func (i *instancematcher) MissingCount() int {
	return int(i.instance.Instances) - len(i.matches)
}
func (i *instancematcher) add(ma *targets.MatchApp) {
	i.matches = append(i.matches, ma)
}
func (i *instancematcher) isUsed() bool {
	if i.instance.Instances == uint32(len(i.matches)) {
		return true
	}
	return false
}
func (r *Resolver) Debugf(format string, args ...interface{}) {
	if !*debug {
		return
	}
	r.Printf(format, args...)
}
func (r *Resolver) Printf(format string, args ...interface{}) {
	prefix := fmt.Sprintf("[RepoID:%d, DeplID:%d, AppID:%d, Bin:%s] ", r.req.DeploymentDescriptor.Application.RepositoryID, r.req.DeploymentDescriptor.ID, r.req.DeploymentDescriptor.Application.ID, r.req.DeploymentDescriptor.Application.Binary)
	fmt.Printf(prefix+format, args...)
}
