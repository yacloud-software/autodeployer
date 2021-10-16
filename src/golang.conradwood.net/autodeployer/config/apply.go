package config

import (
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/autodeployer"
	"golang.conradwood.net/autodeployer/deployments"
)

var (
	// currently applied actions
	applied     []*ActionDep
	debug_apply = flag.Bool("debug_config_apply", false, "enable debug mode for config apply")
)

// mapper to which action is applied to which "Deployed" app
type ActionDep struct {
	actions    []Action
	config     *ApplicationConfig
	deployment *deployments.Deployed
}

// an action specific interface
// actions are created newly for each deployment.
// the lifetime of an action matches that of a deployment.
// we call Apply() when a deployment is started and UnApply() when deployment terminates.
// if an application is terminated and replaced by another one,
// there will be two action objects for a while.
// thus it is guaranteed that unapply() is called on the same object
// on which apply() was called.
type Action interface {
	ID() string // typically deployment.StartupMSG
	Apply() error
	Unapply() error
	String() string
}

// clear list of applied ones
func ClearApplied() {
	ResetPorts()
	applied = make([]*ActionDep, 0)
	fmt.Printf("Cleared list of applied actions\n")
}

// apply config to apps
func Apply() {
	// build a list of stuff that needs doing
	// do this by going through all matchers
	// and verifying if current applied config
	// still refers to the latest deployment
	var proposed []*ActionDep
	if *debug_apply {
		fmt.Printf("Got %d applications in config\n", len(config.Applications))
		fmt.Printf("Got %d deployed applications\n", len(deployments.ActiveDeployments()))
	}

	// build a complete list of how stuff should be applied
	for _, app := range config.Applications {
		if *debug_apply {
			fmt.Printf("Applying config for application: %v\n", app.Matcher)
		}
		var latestDeployment *deployments.Deployed
		for _, d := range deployments.ActiveDeployments() {
			if d.Status != pb.DeploymentStatus_EXECUSER {
				continue
			}
			if !app.matches(d) {
				if *debug_apply {
					fmt.Printf("no match: %v and %s\n", app.Matcher, d.GenericString())
				}
				continue
			}
			if d.Idle {
				continue
			}
			if (latestDeployment == nil) || (latestDeployment.Started.Before(d.Started)) {
				latestDeployment = d
			}
		}
		if latestDeployment != nil {
			ac, err := createActions(app, latestDeployment)
			if err != nil {
				fmt.Printf("Failed to create actions for app %s: %s\n", app.String(), err)
				continue
			}
			ad := ActionDep{config: app, deployment: latestDeployment, actions: ac}
			proposed = append(proposed, &ad)
		}
	}

	// filter away actions which are already applied
	if *debug_apply {
		fmt.Printf("Found %d new proposals in addition to %d existing ones. Filtering out those which are already applied...\n", len(proposed), len(applied))
	}
	var nonapplied []*ActionDep
	for _, r := range proposed {
		alreadyapplied := false
		for _, a := range applied {
			if a.Equals(r) {
				if *debug_apply {
					fmt.Printf("Filtering out %v, because it matches %v\n", r.String(), a.String())
				}
				alreadyapplied = true
				break
			}
		}
		if !alreadyapplied {
			nonapplied = append(nonapplied, r)
		}
	}
	if *debug_apply {
		fmt.Printf("Got %d not-applied proposals. Applying...\n", len(nonapplied))
	}

	// we now got a list of new proposals

	// we now need to build a list of actions that should no longer be applied:
	// (specifically, this means configs to deployments which are no longer pointing to the right deployment)
	var newapplied []*ActionDep
	for _, ad := range applied {
		remove := false
		for _, r := range nonapplied {
			// check if there is a newone to be deployed...
			if r.config == ad.config {
				remove = true
				for _, a := range ad.actions {
					fmt.Printf("Unapplying %s\n", a.String())
					a.Unapply()
				}
			}
		}
		if !remove {
			newapplied = append(newapplied, ad)
		}
	}
	applied = newapplied
	// apply remaining actions in proposed:
	for _, r := range nonapplied {
		fmt.Printf("proposed: %s\n", r.config.String())
		for _, a := range r.actions {
			fmt.Printf("   Applying action: #%s (%s) \n", a.ID(), a.String())
			err := a.Apply()
			if err != nil {
				fmt.Printf("Failed to apply action %s to %s: %s\n", a.String(), r.config.String(), err)
				return
			}
		}
		applied = append(applied, r)
	}
	if *debug_apply { // NOT A DEBUG IF CLAUSE
		printApplied(applied)
	}
}

// given a pointer to a parsed config,
// will create the action structs
// if one adds new actions, these need to be included here
func createActions(ac *ApplicationConfig, du *deployments.Deployed) ([]Action, error) {
	var res []Action
	if len(ac.ActionPorts) > 0 {
		acp, err := createActionPorts(ac, du)
		if err != nil {
			return nil, err
		}
		res = append(res, acp...)
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("ApplicationConfig %s has no actions", ac.String())
	}
	return res, nil

}

func createActionPorts(ac *ApplicationConfig, du *deployments.Deployed) ([]Action, error) {
	var res []Action
	for _, ap := range ac.ActionPorts {
		action, err := NewPortAction(ap, du)
		if err != nil {
			return nil, err
		}
		res = append(res, action)
	}
	return res, nil
}

func (ad *ActionDep) String() string {
	return fmt.Sprintf("%s to %s", ad.config.String(), ad.deployment.String())
}

// return true only if all fields are identical
func (ad *ActionDep) Equals(cmp *ActionDep) bool {
	if ad.config != cmp.config {
		return false
	}
	if ad.deployment != cmp.deployment {
		return false
	}
	if len(ad.actions) != len(cmp.actions) {
		return false
	}
	return true
}

func printApplied(ads []*ActionDep) {
	fmt.Printf("===============================\n")
	fmt.Printf("%d applied actions:\n", len(ads))
	for _, a := range ads {
		fmt.Printf("   %s -> %s\n", a.config.String(), a.deployment.String())
	}
	fmt.Printf("===============================\n")
}
