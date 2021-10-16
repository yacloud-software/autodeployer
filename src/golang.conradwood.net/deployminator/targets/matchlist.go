package targets

import (
	ad "golang.conradwood.net/apis/autodeployer"
	"strings"
)

type MatchList struct {
	parent *MatchList
	apps   []*MatchApp
}
type MatchApp struct {
	target *Target
	app    *ad.DeployedApp
}

// returns any machine groups which are configured and match this target
func (ma *MatchApp) MatchingMachineGroups() []string {
	cfg := ma.ConfiguredMachineGroups()
	tm := ma.TargetMachineGroups()
	var res []string
	for _, t := range tm {
		found := false
		for _, c := range cfg {
			if c == t {
				found = true
				break
			}
		}
		if found {
			res = append(res, t)
		}
	}
	return res
}

// get the machine groups for this target
func (ma *MatchApp) TargetMachineGroups() []string {
	return ma.target.machinegroups
}

// get the machine groups configured for this application
func (ma *MatchApp) ConfiguredMachineGroups() []string {
	s := ma.app.Deployment.AppReference.AppDef.Machines
	if s == "" { // if machinegroup is not configured, it defaults to 'worker'
		return []string{"worker"}
	}
	res := strings.Split(s, ",")
	return res
}

func (ma *MatchApp) HostMachineGroups() []string {
	s := ma.target.machinegroups
	if len(s) == 0 { // if machinegroup is not configured, it defaults to 'worker'
		return []string{"worker"}
	}
	return s
}
func (ma *MatchApp) Host() string {
	return ma.target.address
}
func (ma *MatchApp) App() *ad.DeployedApp {
	return ma.app
}
func (ml *MatchList) add(t *Target, app *ad.DeployedApp) {
	ml.apps = append(ml.apps, &MatchApp{target: t, app: app})
}
func (ml *MatchList) Remove(ma *MatchApp) {
	var res []*MatchApp
	for _, a := range ml.apps {
		if a != ma {
			res = append(res, a)
		}
	}
	ml.apps = res
}

func (ml *MatchList) FilterByAppBinary(binary string) *MatchList {
	res := &MatchList{parent: ml}
	for _, ma := range ml.apps {
		if ma.app.Deployment.Binary == binary {
			res.apps = append(res.apps, ma)
		}
	}
	return res
}

func (ml *MatchList) FilterByAppRepo(repo uint64) *MatchList {
	res := &MatchList{parent: ml}
	for _, ma := range ml.apps {
		if ma.app.Deployment.RepositoryID == repo {
			res.apps = append(res.apps, ma)
		}
	}
	return res
}

/* do not filter by url - it is modifed to replace vars like ${BUILDID}
func (ml *MatchList) FilterByAppURL(url string) *MatchList {
	res := &MatchList{parent: ml}
	for _, ma := range ml.apps {
		if ma.app.Deployment.DownloadURL == url {
			res.apps = append(res.apps, ma)
		}
	}
	return res
}
*/
func (ml *MatchList) FilterByAppBuild(build uint64) *MatchList {
	res := &MatchList{parent: ml}
	for _, ma := range ml.apps {
		if ma.app.Deployment.BuildID == build {
			res.apps = append(res.apps, ma)
		}
	}
	return res
}
func (ml *MatchList) Apps() []*MatchApp {
	return ml.apps
}
