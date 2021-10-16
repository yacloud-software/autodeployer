package main

// responsible for shutting down services which are no longer needed
// (rather: marked for shutdown)
import (
	"flag"
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	rg "golang.conradwood.net/apis/registry"
)

var (
	debug_cond_running = flag.Bool("debug_cond_running", false, "enable debug for the running condition before stopping")
)

type StopperRunningCondition struct {
	startupid  string
	host       *rg.ServiceAddress
	minruntime int
}

func (s *StopperRunningCondition) String() string {
	return fmt.Sprintf("App %s on %s running for %ds?", s.startupid, s.host, s.minruntime)
}
func (s *StopperRunningCondition) eval(sr *stopRequest) (int, error) {
	if *debug_cond_running {
		fmt.Printf("Evaluating cond for %s\n", sr.String())
	}
	app, err := GetDeployInfo(s.host.Host, s.startupid)
	if err != nil {
		return 1, err // error
	}
	if app == nil {
		return 2, nil // false
	}
	if app.Deployment.RuntimeSeconds >= uint64(s.minruntime) {
		return 3, nil // true
	}
	return 1, nil // inconclusive
}

func GetDeployInfo(host string, startupid string) (*ad.DeployedApp, error) {
	ir, err := GetDeployments(host)
	if err != nil {
		return nil, err
	}
	for _, app := range ir.Apps {
		if app.ID == startupid {
			return app, nil
		}
	}
	return nil, nil
}
