package suggest

// this finds applications with too few instances and suggests fixes
// (e.g. start of new ones)
import (
	"flag"
	"fmt"

	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/useroverride"
)

var (
	missing_deployers_count_zero = flag.Bool("count_missing_deployers_as_zero", true, "if true assumes that machinegroups without deployers have 0 instances (but continue with suggestions")
)

type fixMissing struct {
	suggestion *Suggestion
}

func NewFixMissing(c *Suggestion) *fixMissing {
	return &fixMissing{suggestion: c}
}

func (f *fixMissing) Run() {
	iter := f.suggestion.config.AppIterator()
	if *debugSuggest {
		fmt.Printf(" FixMissing(): (%d apps)\n", len(iter))
	}
	appids := make(map[uint64]bool) // keep a record which apps (by id) we counted instances for
	for _, ai := range iter {
		appids[ai.App.ID] = true
		if !ai.App.AlwaysOn {
			continue
		}
		if useroverride.GetMetaData(ai.App).UserDisabled() {
			continue
		}
		if ai.App.InstancesMeansPerAutodeployer {
			// TODO: monitor per-instance count apps
			continue
		}
		actual := CountInstances(f.suggestion.ProjectedDeployments(), ai.App)
		wanted := int(ai.App.Instances)

		if *debugSuggest {
			fmt.Printf("   %d %s \n", ai.App.ID, ai.App.Binary)
			fmt.Printf("      Actual: %d | Wanted: %d\n", actual, wanted)
		}
		if actual == wanted {
			continue
		}
		for actual < wanted {
			m := ai.App.Machines
			if m == "" {
				m = "worker"
			}
			deployers := f.suggestion.config.Deployers.ByGroup(m)
			if len(deployers.Targets) == 0 {
				fmt.Printf("#1 Failed to find any deployers for type \"%s\"\n", m)
				f.suggestion.AddMissingDeployer(m)
				if !*missing_deployers_count_zero {
					return
				}
			}
			deployers = f.suggestion.ByLeastInstances(deployers, ai.App)
			if *debugSuggest {
				fmt.Printf("  deploy on (byLeast instances): %v\n", deployers)
			}
			if len(deployers.Targets) == 0 {
				fmt.Printf("#1 Failed to find least-instance deployer for \"%s\"\n", ai.App.Binary)
				if *missing_deployers_count_zero {
					break
				} else {
					return
				}
			}
			f.suggestion.AddStart(&StartApp{Host: deployers.Targets[0].Host, App: ai.App})
			actual++
		}
		for actual > wanted {
			deployers := f.suggestion.config.Deployers.ByGroup(ai.App.Machines)
			if len(deployers.Targets) == 0 {
				fmt.Printf("#2 Failed to find any deployers for type \"%s\"\n", ai.App.Machines)
				if *missing_deployers_count_zero {
					break
				} else {
					return
				}
			}
			deployers = f.suggestion.ByMostInstances(deployers, ai.App)
			if len(deployers.Targets) == 0 {
				fmt.Printf("#2 Failed to find most-instance deployer for %s\n", ai.App.Binary)
				if *missing_deployers_count_zero {
					break
				} else {
					return
				}
			}
			apps := f.suggestion.GetInstancesOnTarget(deployers.Targets[0], ai.App)
			if len(apps) != 0 {
				f.suggestion.AddStop(&StopApp{Host: deployers.Targets[0].Host, App: apps[0]})
			}
			actual--
		}
	}

	// it is possible that there are apps which we did not count.
	// that is those with an app.ID which aren't in the config.
	// these 'stray' ones should be stopped.
	for _, dl := range f.suggestion.ProjectedDeployments().Deployments {
		for _, gd := range dl.Apps {
			for _, app := range gd.Applications {
				if appids[app.ID] {
					continue
				}
				f.suggestion.AddStop(&StopApp{Host: dl.Host, App: app})
				if *debugSuggest {
					fmt.Printf("Stray App: %d %s on %s\n", app.ID, app.Binary, dl.Host)
				}
			}

		}
	}
	if *debugSuggest {
		fmt.Printf("done fixmissing()\n")
	}
}

// how many instances do we have running of a given app?
// to be counted as a match, a running instance must match on
// the applicationid
func CountInstances(dl *pb.DeploymentList, app *pb.ApplicationDefinition) int {
	res := 0
	for _, d := range dl.Deployments {
		for _, gd := range d.Apps {
			for _, a := range gd.Applications {
				if a.ID == app.ID {
					res++
				}
			}
		}

	}
	return res
}
