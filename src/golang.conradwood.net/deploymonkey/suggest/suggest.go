package suggest

// this package works out a series of 'fixes' for
// the datacenter based on the current Configuration
// and the actual state.
// it is intentionally written as a package so it
// can be invoked by commandline by a human.
// Eventually, over time, when trust in this package
// grows, bugs are ironed out, it *may* be useful
// to run this periodically and unattended

import (
	"bytes"
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/config"
	"sort"
)

var (
	debugSuggest = flag.Bool("debug_suggest", false, "enable debug of suggest code")
)

// a list of fixes and state
type Suggestion struct {
	deployments      *pb.DeploymentList
	config           *config.Config
	Starts           []*StartApp
	Stops            []*StopApp
	missingDeployers []string
}

type StartApp struct {
	Host string
	App  *pb.ApplicationDefinition
}
type StopApp struct {
	Host string
	App  *pb.ApplicationDefinition
}

func (s *StartApp) DeployRequest() *pb.DeployAppRequest {
	return &pb.DeployAppRequest{AppID: s.App.ID, Host: s.Host}
}
func (s *StopApp) UndeployRequest() *pb.UndeployAppRequest {
	return &pb.UndeployAppRequest{DeploymentID: s.App.DeploymentID, Host: s.Host}
}
func (s *Suggestion) MissingDeployers() []string {
	return s.missingDeployers
}
func (s *Suggestion) AddMissingDeployer(m string) {
	for _, ms := range s.missingDeployers {
		if ms == m {
			return
		}
	}
	s.missingDeployers = append(s.missingDeployers, m)
}
func (s *Suggestion) AddStart(a *StartApp) {
	s.Starts = append(s.Starts, a)
}
func (s *Suggestion) AddStop(a *StopApp) {
	s.Stops = append(s.Stops, a)
}

func (s *Suggestion) Deployments() *pb.DeploymentList {
	return s.deployments
}

// deployments + fixes (start/stop) applied
func (s *Suggestion) ProjectedDeployments() *pb.DeploymentList {
	res := &pb.DeploymentList{}
	// add all those which are not marked as stopped
	for _, d := range s.deployments.Deployments {
		// ranging over a list of Host->[]apps
		nd := &pb.Deployment{Host: d.Host}
		res.Deployments = append(res.Deployments, nd)
		for _, gd := range d.Apps {
			ngd := &pb.GroupDefinitionRequest{Namespace: gd.Namespace, GroupID: gd.GroupID}
			nd.Apps = append(nd.Apps, ngd)
			for _, ap := range gd.Applications {
				if !s.stopped(nd.Host, ap) {
					ngd.Applications = append(ngd.Applications, ap)
				}
			}
		}
	}
	// add all those which are marked as "to be started"
	for _, ac := range s.Starts {
		apps := []*pb.ApplicationDefinition{ac.App}
		gdr := pb.GroupDefinitionRequest{Applications: apps}
		d := &pb.Deployment{Host: ac.Host, Apps: []*pb.GroupDefinitionRequest{&gdr}}
		res.Deployments = append(res.Deployments, d)
	}

	return res
}

func (s *Suggestion) Count() int {
	return len(s.Stops) + len(s.Starts)
}

// returns true if we have a stop request for this tuple
func (s *Suggestion) stopped(host string, app *pb.ApplicationDefinition) bool {
	for _, stop := range s.Stops {
		if (stop.Host == host) && (app.ID == stop.App.ID) {
			return true
		}
	}
	return false
}

// return true if both suggestions are exactly equal
func (s *Suggestion) Equals(s1 *Suggestion) bool {
	if s1 == nil || s == nil {
		return false
	}
	// KISS
	if s.String() == s1.String() {
		return true
	}
	return false
}

// config -> deploymonkeys status in database
// dl -> what's ACTUALLY deployed
func Analyse(conf *config.Config, dl *pb.DeploymentList) (*Suggestion, error) {
	res := &Suggestion{deployments: dl, config: conf}
	if *debugSuggest { // NOT A DEBUG IF CLAUSE
		fmt.Printf("** Analysing config for suggestions **\n")
		fmt.Printf(" Config:\n")
		for d, ai := range conf.AppIterator() {
			fmt.Printf(" %3d.  %s [%s]\n", d+1, ai.App.Binary, ai.App.DeploymentID)
		}
		fmt.Printf(" Deployments:\n")
		for _, d := range dl.Deployments {
			for _, g := range d.Apps {
				for _, ai := range g.Applications {
					fmt.Printf("  %s [%s]\n", ai.Binary, ai.DeploymentID)
				}
			}
		}
	}
	f := NewFixMissing(res)
	f.Run()
	if *debugSuggest {
		fmt.Println(res.String())
	}
	return res, nil
}

func (ac *StopApp) String() string {
	return fmt.Sprintf("Stop App #%d (%d/%s) on %s (deploymentid %s)\n", ac.App.ID, ac.App.RepositoryID, ac.App.Binary, ac.Host, ac.App.DeploymentID)

}
func (ac *StartApp) String() string {
	return fmt.Sprintf("Start App #%d (%d/%s) on %s\n", ac.App.ID, ac.App.RepositoryID, ac.App.Binary, ac.Host)
}

func (s *Suggestion) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("Suggestions:\n")
	for _, ac := range s.Starts {
		buffer.WriteString(fmt.Sprintf("Start App #%d (%d/%s) on %s\n", ac.App.ID, ac.App.RepositoryID, ac.App.Binary, ac.Host))
	}
	for _, ac := range s.Stops {
		buffer.WriteString(fmt.Sprintf("Stop App #%d (%d/%s) on %s (deploymentid %s)\n", ac.App.ID, ac.App.RepositoryID, ac.App.Binary, ac.Host, ac.App.DeploymentID))
	}

	return buffer.String()

}

func (s *Suggestion) countInstancesOnTarget(d *pb.Deployer, app *pb.ApplicationDefinition) int {
	res := 0
	for _, dt := range s.ProjectedDeployments().Deployments {
		if dt.Host != d.Host {
			continue
		}
		for _, gd := range dt.Apps {
			for _, ap := range gd.Applications {
				if app.ID == ap.ID {
					res++
				}
			}
		}

	}
	return res
}

func (s *Suggestion) GetInstancesOnTarget(d *pb.Deployer, app *pb.ApplicationDefinition) []*pb.ApplicationDefinition {
	var res []*pb.ApplicationDefinition
	for _, dt := range s.ProjectedDeployments().Deployments {
		if dt.Host != d.Host {
			continue
		}
		for _, gd := range dt.Apps {
			for _, ap := range gd.Applications {
				if app.ID == ap.ID {
					res = append(res, ap)
				}
			}
		}

	}
	return res
}

// how many deployments on this target?
func (s *Suggestion) countDeploymentsOnTarget(d *pb.Deployer) int {
	res := 0
	for _, dt := range s.ProjectedDeployments().Deployments {
		if dt.Host != d.Host {
			continue
		}
		res = res + len(dt.Apps)

	}
	return res

}

// deployers who are running the least amount of instances of this app, sorted by number-of-total-deployments
func (s *Suggestion) ByLeastInstances(deployers *config.Deployers, app *pb.ApplicationDefinition) *config.Deployers {
	res := &config.Deployers{}

	// find any with 0 deployments
	for _, target := range deployers.Targets {
		l := s.countDeploymentsOnTarget(target)
		if *debugSuggest {
			fmt.Printf("Deployments on %s: %d\n", target, l)
		}
		if l == 0 {
			res.Targets = append(res.Targets, target)
		}
	}
	// if we have targets with 0 installed applications, we don't need to bother
	// checking for least deployment targets, because they are not going to be less than zero
	// in fact we want to prefer those over others
	if len(res.Targets) != 0 {
		return res
	}
	ct := -1
	for _, dt := range s.ProjectedDeployments().Deployments {
		dl := deployers.ByIP(dt.Host)
		if len(dl.Targets) == 0 {
			continue
		}
		b := dl.Targets[0]

		// do not add the same target twice
		found := false
		for _, x := range res.Targets {
			if b == x {
				found = true
			}
		}
		if found {
			continue
		}

		f := s.countInstancesOnTarget(b, app)

		if (ct == -1) || (ct == f) {
			res.Targets = append(res.Targets, b)
			ct = f
			continue
		}
		if f == ct {
			res.Targets = append(res.Targets, b)
			continue
		}
		if f < ct {
			res.Targets = []*pb.Deployer{b}
			ct = f
			continue
		}
	}

	// if we have no "least count" but we have deployers to start with, then pick first one
	if (len(res.Targets) == 0) && (len(deployers.Targets) != 0) {
		fmt.Printf("No least count - using all deployers\n")
		res.Targets = append(res.Targets, deployers.Targets[0])
	}
	// sort it so that the target with the least deployments is first.
	sort.Slice(res.Targets, func(i, j int) bool {
		a := s.countDeploymentsOnTarget(res.Targets[i])
		b := s.countDeploymentsOnTarget(res.Targets[j])
		return a < b
	})
	return res
}

// deployers who are running the highest amount of instances of this app
func (s *Suggestion) ByMostInstances(deployers *config.Deployers, app *pb.ApplicationDefinition) *config.Deployers {
	res := &config.Deployers{}
	ct := -1
	for _, dt := range s.ProjectedDeployments().Deployments {
		dl := deployers.ByIP(dt.Host)
		if len(dl.Targets) == 0 {
			continue
		}
		b := dl.Targets[0]
		f := s.countInstancesOnTarget(b, app)
		if (ct == -1) || (ct == f) {
			res.Targets = append(res.Targets, b)
			ct = f
			continue
		}
		if f == ct {
			res.Targets = append(res.Targets, b)
			continue
		}
		if f > ct {
			res.Targets = []*pb.Deployer{b}
			ct = f
			continue
		}
	}
	return res
}
