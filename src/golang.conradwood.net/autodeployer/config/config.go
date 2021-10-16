package config

// this package enables system administrators to configure
// "special" handling for packages on the local machines

import (
	"flag"
	"fmt"
	"golang.conradwood.net/autodeployer/deployments"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

var (
	config        *AutoDeployerConfig
	lastRead      time.Time
	sampleConfig  = flag.Bool("sample_config", false, "if true prints out a sample config file")
	sleepinterval = flag.Int("apply_action_interval", 15, "interval in `seconds` to re-apply the actions to deployments")
)

// this is the local autodeployer config

type AutoDeployerConfig struct {
	NFT_Templates []string
	Applications  []*ApplicationConfig
}
type ApplicationConfig struct {
	Label       string
	Matcher     *ApplicationMatcher
	ActionPorts []*ApplicationPort
}

func (a *ApplicationConfig) String() string {
	if a.Matcher == nil {
		return "unnamed config"
	}
	return fmt.Sprintf("%s/%s/%d/%s", a.Matcher.Namespace, a.Matcher.Groupname, a.Matcher.RepositoryID, a.Matcher.Binary)
}

type ApplicationMatcher struct {
	RepositoryID uint64
	Binary       string
	Groupname    string
	Namespace    string
}
type ApplicationPort struct {
	PortIndex  int
	PublicPort int
}

// not having a config is NOT an error
// having a config, but not parsing it, is an error
func Start() error {
	if *sampleConfig {
		printSampleConfig()
		os.Exit(0)
	}
	fname := "/etc/cnw/autodeployer/config.yaml"
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		return nil
	}
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return fmt.Errorf("config: File %s exists, but reading it failed: %s", fname, err)
	}
	lastRead = time.Now()
	cfg := AutoDeployerConfig{}
	err = yaml.UnmarshalStrict(b, &cfg)
	if err != nil {
		return fmt.Errorf("config: File %s exists, but parsing it failed: %s", fname, err)
	}
	config = &cfg
	ResetPorts()
	if config != nil {
		// can't remember the syntax for ticker...
		go func() {
			for {
				reread()
				Apply()
				time.Sleep(time.Duration(*sleepinterval) * time.Second)
			}
		}()
	}
	return nil
}

// periodically check if the config file needs reloading
func reread() error {
	fname := "/etc/cnw/autodeployer/config.yaml"
	s, err := os.Stat(fname)
	if os.IsNotExist(err) {
		return nil
	}
	t := s.ModTime()
	if t.Before(lastRead) {
		return nil
	}
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return fmt.Errorf("config: File %s exists, but reading it failed: %s", fname, err)
	}
	lastRead = time.Now()
	cfg := AutoDeployerConfig{}
	err = yaml.UnmarshalStrict(b, &cfg)
	if err != nil {
		return fmt.Errorf("config: File %s exists, but parsing it failed: %s", fname, err)
	}
	config = &cfg
	fmt.Printf("Config %s re-read\n", fname)
	return nil
}

// goes through list of deployed stuff and apply
// the actions to applications (e.g. redirect ports)
// it will 'ignore' applications which have been running for less
// than 15 seconds (treat those as 'not' deployed/non-existent)
func AppStarted(d *deployments.Deployed) {
	if config == nil {
		return
	}
	fmt.Printf("[config] - applying local stuff because %s started up\n", d.GenericString())
	Apply()
	return
}
func AppStopped() {
	if config == nil {
		return
	}
}

// print a sample config to display the yamlsyntax
func printSampleConfig() {
	cfg := AutoDeployerConfig{NFT_Templates: []string{nf_templ1, nf_templ2}}

	ac := ApplicationConfig{
		Matcher:     &ApplicationMatcher{RepositoryID: 5, Binary: "testbinary1", Namespace: "namespace1"},
		ActionPorts: []*ApplicationPort{&ApplicationPort{1, 80}},
	}
	cfg.Applications = append(cfg.Applications, &ac)
	ac = ApplicationConfig{
		Matcher:     &ApplicationMatcher{Groupname: "mygroup", RepositoryID: 6, Binary: "testbinary2", Namespace: "namespace2"},
		ActionPorts: []*ApplicationPort{&ApplicationPort{1, 10}, &ApplicationPort{2, 22}},
	}
	cfg.Applications = append(cfg.Applications, &ac)
	b, _ := yaml.Marshal(&cfg)
	fmt.Println(string(b))
}

// get a config for a deployed app (or nil)
func getConfig(d *deployments.Deployed) *ApplicationConfig {
	if config == nil {
		return nil
	}
	for _, app := range config.Applications {
		if !app.matches(d) {
			continue
		}
		return app
	}
	return nil
}

// shortcut to Matcher.matches()
func (am *ApplicationConfig) matches(d *deployments.Deployed) bool {
	if am.Matcher == nil {
		return false
	}
	return am.Matcher.matches(d)
}

// check if a deployed application matches a config entry
func (am *ApplicationMatcher) matches(d *deployments.Deployed) bool {
	if (am.RepositoryID != 0) && (am.RepositoryID != d.RepositoryID()) {
		return false
	}
	if (am.Namespace != "") && (am.Namespace != d.Namespace()) {
		return false
	}
	if (am.Binary != "") && (am.Binary != d.Binary()) {
		return false
	}
	if (am.Groupname != "") && (am.Groupname != d.Groupname()) {
		return false
	}
	return true
}
