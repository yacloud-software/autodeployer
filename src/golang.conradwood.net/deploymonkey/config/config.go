package config

import (
	"flag"
	"fmt"
	"golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/go-easyops/authremote"
)

var (
	debug_config = flag.Bool("debug_config", false, "Debug config code")
)

/************************************************
* Main object: "Config"
*************************************************/
type Config struct {
	gc        []*pb.GroupConfig
	Deployers *Deployers
}

func (c *Config) Namespaces() []string {
	if c.gc == nil {
		return []string{}
	}
	var res []string
	for _, gsc := range c.gc {
		ns := gsc.Group.NameSpace
		found := false
		for _, r := range res {
			if r == ns {
				found = true
				break
			}
		}
		if !found {
			res = append(res, ns)
		}
	}
	return res
}

func (c *Config) Groups(Namespace string) []*pb.GroupDef {
	var res []*pb.GroupDef
	for _, gcs := range c.gc {
		g := gcs.Group
		if (Namespace == "") || (g.NameSpace == Namespace) {
			res = append(res, g)
		}
	}
	return res
}

// groupname is equivalent to groupid (apparently)
func (c *Config) Apps(Namespace string, Groupname string) []*pb.ApplicationDefinition {
	var res []*pb.ApplicationDefinition
	for _, gcs := range c.gc {
		g := gcs.Group
		if (Namespace != "") && (g.NameSpace != Namespace) {
			continue
		}
		for _, x := range gcs.Applications {
			res = append(res, x)
		}
	}
	return res
}

// true if this group is in the deployer
func ContainsGroup(a []string, b string) bool {
	for _, x := range a {
		if x == b {
			return true
		}
	}
	return false
}

// true if both arrays contain same groups
func SameGroups(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, x := range a {
		found := false
		for _, y := range b {
			if y == x {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// returns true if both configs have the same set of autodeployers (irrespective of which applications are installed on each)
func (c *Config) HasSameAutodeployers(cc *Config) bool {
	if c.Deployers.Targets == nil || cc.Deployers.Targets == nil {
		return false
	}
	if len(c.Deployers.Targets) != len(cc.Deployers.Targets) {
		return false
	}
	for _, t := range c.Deployers.Targets {
		found := false
		for _, d1 := range cc.Deployers.Targets {
			if d1.Host == t.Host && SameGroups(d1.Machinegroup, t.Machinegroup) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for _, t := range cc.Deployers.Targets {
		found := false
		for _, d1 := range c.Deployers.Targets {
			if d1.Host == t.Host && SameGroups(d1.Machinegroup, t.Machinegroup) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

/************************************************
* Helper: App-iterator (to iterate over each app)
*************************************************/
type AppIterator struct {
	Group *pb.GroupDef
	App   *pb.ApplicationDefinition
}

func (c *Config) AppIterator() []*AppIterator {
	var res []*AppIterator
	for _, ns := range c.Namespaces() {
		for _, g := range c.Groups(ns) {
			for _, a := range c.Apps(ns, g.GroupID) {
				ai := &AppIterator{Group: g, App: a}
				res = append(res, ai)
			}
		}
	}
	return res
}

/************************************************
* Deployer object (listing possible targets)
*************************************************/
type Deployers struct {
	config  *Config
	Targets []*pb.Deployer
}

// get all deployers which list this group
func (d *Deployers) ByGroup(group string) *Deployers {
	if group == "" {
		group = "worker"
	}
	res := &Deployers{config: d.config}
	for _, t := range d.Targets {
		if ContainsGroup(t.Machinegroup, group) {
			res.Targets = append(res.Targets, t)
		}
	}
	if *debug_config {
		fmt.Printf("Group %s has %d targets\n", group, len(d.Targets))
	}
	return res
}

func (d *Deployers) ByIP(ip string) *Deployers {
	res := &Deployers{config: d.config}
	for _, d := range d.Targets {
		if d.Host == ip {
			res.Targets = append(res.Targets, d)
		}
	}
	return res

}

/************************************************
* Get the entire config
*************************************************/
func GetConfig(depl pb.DeployMonkeyClient) (*Config, error) {
	var err error
	var res = &Config{}
	pbc, err := depl.GetConfig(authremote.Context(), &common.Void{})
	if err != nil {
		return nil, err
	}
	printConfig(pbc)
	res.gc = pbc.GroupConfigs
	res.Deployers = &Deployers{config: res, Targets: pbc.Deployers.Deployers}
	res.Deployers.DebugPrintTargets()
	return res, nil
}
func printConfig(c *pb.Config) {
	if !*debug_config { // NOT A DEBUG IF CLAUSE
		return
	}
	fmt.Printf("Deployments:\n")
	for _, g := range c.GroupConfigs {
		fmt.Printf("   Group %s: %d applications\n", g.Group.NameSpace, len(g.Applications))
		for _, a := range g.Applications {
			fmt.Printf("      App: %s [%s]\n", a.Binary, a.DeploymentID)
		}
	}
}

func (d *Deployers) DebugPrintTargets() {
	if !*debug_config { // NOT A DEBUG IF CLAUSE
		return
	}
	fmt.Printf("*** Config - got %d targets ***\n", len(d.Targets))
	for _, t := range d.Targets {
		fmt.Printf("config() target: %s %s\n", t.Host, t.Machinegroup)
	}
}
