package main

import (
	"flag"
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	dm "golang.conradwood.net/apis/deploymonkey"
	rpb "golang.conradwood.net/apis/registry"
	"golang.conradwood.net/deploymonkey/config"
	"golang.conradwood.net/deploymonkey/scheduler"
	"golang.conradwood.net/deploymonkey/suggest"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/prometheus"
	"golang.conradwood.net/go-easyops/tokens"
	"strings"
	"sync"
	"time"
)

const (
	MAX_QUERY_FAILURES = 5 // given that many consecutive failures, this deployer is considered broken
)

var (
	enable_auto_suggest_ctr = flag.Bool("enable_auto_suggest_ctr", true, "If enabled, automatically runs the suggest code after the scanner has ran (thus -enable_scanner must be true as well)")
	scan_interval           = flag.Int("scan_interval", 30, "interval in `seconds` between scans of autodeployers")
	autodeployers           []*AutoDeployer
	scan_in_progress        sync.Mutex
	suggestioncount         = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "deploymonkey_suggestions",
			Help: "V=1 UNIT=none DESC=number of suggestions deploymonkey currently has to offer",
		},
		[]string{},
	)
	suggestion_missing_deployers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "deploymonkey_suggestions_missing_deployers_count",
			Help: "V=1 UNIT=none DESC=number of machinegroups which are referenced in deployments but not available",
		},
		[]string{},
	)
	machinegroup_deployers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "deploymonkey_machinegroup_deployers_count",
			Help: "V=1 UNIT=none DESC=number of autodeployers serving each machinegroups (which is referenced by deployments or served)",
		},
		[]string{"machinegroup"},
	)
)

type AutoDeployer struct {
	IP            string
	Port          int32
	Group         []string
	LastSeen      time.Time
	queryFailures int
	Apps          []*ad.DeployedApp
	lock          sync.Mutex
	Broken        bool
	Available     bool
	tmpFound      bool // we use this to mark it temporarily as "found" in the registry, all those with false will become unavailbable - see ScanAutodeployers()
}

/*
 this starts a timer which periodically scans all autodeployers
 we maintain a copy of the deployments in-cache
*/
func StartScanner() {
	if *debug {
		fmt.Printf("Starting scanner...\n")
	}
	var meClient dm.DeployMonkeyClient
	go func() {
		meConn := client.Connect("deploymonkey.DeployMonkey") // connect to service instead of deploymentpath
		meClient = dm.NewDeployMonkeyClient(meConn)
	}()
	prometheus.MustRegister(suggestioncount, suggestion_missing_deployers, machinegroup_deployers)
	ticker := time.NewTicker(time.Duration(*scan_interval) * time.Second)
	go func() {
		ScanAutodeployers()
		var prevConfig *config.Config
		var prevSuggest *suggest.Suggestion
		for _ = range ticker.C {
			if *debug {
				fmt.Printf("Scanning autodeployers for applications...\n")
			}
			ScanAutodeployers()
			if *debug { // NOT A DEBUG IF CLAUSE
				PrintAutodeployers()
			}
			setMachineGroupGauge()
			if (!*enable_auto_suggest_ctr) || (meClient == nil) {
				if *debug {
					fmt.Printf("Weird thing happened, cannot scan\n")
				}
				continue
			}
			dl, err := INT_GetDeploymentsFromCache(tokens.ContextWithToken())
			if err != nil {
				fmt.Printf("scanner: Failed to get deployments from cache: %s\n", err)
				continue
			}
			cfg, err := config.GetConfig(meClient)
			if err != nil {
				fmt.Printf("scanner: Failed to get config: %s\n", err)
				continue
			}
			//fmt.Printf("Config: %v\n, Deployments: %v\n", cfg, dl)
			sug, err := suggest.Analyse(cfg, dl)
			if err != nil {
				fmt.Printf("scanner: analysis failed: %s\n", err)
				continue
			}
			for _, m := range sug.MissingDeployers() {
				machinegroup_deployers.With(prometheus.Labels{"machinegroup": m}).Set(0)
			}
			suggestion_missing_deployers.Set(float64(len(sug.MissingDeployers())))
			ct := sug.Count()
			if *debug {
				fmt.Printf("Suggestion Count: %d\n", ct)
			}
			suggestioncount.With(prometheus.Labels{}).Set(float64(ct))
			if (prevConfig != nil) && (!prevConfig.HasSameAutodeployers(cfg)) {
				fmt.Printf("%d deployments\n", len(cfg.Deployers.Targets))
				scheduler.DefaultScheduler.ConfigChanged()
			}
			prevConfig = cfg
			if (prevSuggest != nil) && (!prevSuggest.Equals(sug)) {
				fmt.Printf("Found %d suggestions\n", ct)
				scheduler.DefaultScheduler.SuggestionsChanged(sug)
			}
			prevSuggest = sug

		}
	}()
}

func ScanAutodeployersTest() {
	err := ScanAutodeployers()
	if err != nil {
		fmt.Printf("Scanning for Autodeployers failed: %s\n", err)
	}
	PrintAutodeployers()
	StartScanner()
	for {
		time.Sleep(1)
	}
}
func ScanAutodeployers() error {
	scan_in_progress.Lock()
	defer scan_in_progress.Unlock()
	sas, err := GetDeployers() // get deployers from registry
	if err != nil {
		return err
	}
	// first of, we mark all those in our list, but not in the registry
	// as broken
	for _, a := range autodeployers {
		a.tmpFound = false
	}
	for _, sa := range sas {
		ad := getDeployerByAddress(sa.Host, sa.Port)
		if ad != nil {
			ad.tmpFound = true
		}
	}
	for _, a := range autodeployers {
		if a.tmpFound == false {
			a.Available = false
		} else {
			a.Available = true
		}
	}

	// iterate over list of deployer as registry has reported
	for _, sa := range sas {
		err = ScanAutodeployer(sa)
		if err != nil {
			fmt.Printf("Error scanning %s: %s\n", sa.Host, err)
			continue
		}
	}
	for _, sa := range sas {
		setBrokennessOnDeployers(sa)
	}
	return nil
}

func setMachineGroupGauge() {
	ctr := make(map[string]int)
	for _, ad := range autodeployers {
		for _, g := range ad.Group {
			l := ctr[g]
			l++
			ctr[g] = l
		}
	}
	for m, c := range ctr {
		machinegroup_deployers.With(prometheus.Labels{"machinegroup": m}).Set(float64(c))
	}
}

func setBrokennessOnDeployers(sa *rpb.ServiceAddress) {
	nt := getDeployerByAddress(sa.Host, sa.Port)
	if nt == nil {
		return
	}
	nt.Broken = (nt.queryFailures > MAX_QUERY_FAILURES)
}

func ScanAutodeployer(sa *rpb.ServiceAddress) error {
	if *debug {
		fmt.Printf("Scanning Host %s\n", sa.Host)
	}
	nt := getDeployerByAddress(sa.Host, sa.Port)
	if nt == nil {
		nt = &AutoDeployer{IP: sa.Host, Port: sa.Port, Available: true}
	}
	nt.lock.Lock()
	defer nt.lock.Unlock()
	// get machinegroup of autodeployer
	conn, err := DialService(sa)
	if err != nil {
		s := fmt.Sprintf("Failed to connect to service %s:%d", sa.Host, sa.Port)
		if *debug {
			fmt.Println(s)
		}
		incFailure(sa.Host, sa.Port)
		return fmt.Errorf("%s", s)
	}
	adc := ad.NewAutoDeployerClient(conn)
	req := &ad.MachineInfoRequest{}
	ctx := tokens.ContextWithToken()
	mir, err := adc.GetMachineInfo(ctx, req)
	if err != nil {
		s := fmt.Sprintf("Failed to get machineinfo on %s: %s", sa.Host, err)
		fmt.Println(s)
		incFailure(sa.Host, sa.Port)
		conn.Close()
		return fmt.Errorf("%s", s)
	}
	da, err := adc.GetDeployments(ctx, &ad.InfoRequest{})
	if err != nil {
		incFailure(sa.Host, sa.Port)
		conn.Close()
		return fmt.Errorf("Failed to get deployment info on %s: %s\n", sa.Host, err)
	}
	conn.Close()
	nt.queryFailures = 0
	nt.Group = mir.MachineGroup
	nt.LastSeen = time.Now()
	nt.Apps = da.Apps
	if *debug {
		fmt.Printf("Found %d apps on %s:%d\n", len(nt.Apps), sa.Host, sa.Port)
	}
	updateAutoDeployer(nt)
	return nil
}

func PrintAutodeployers() {
	for _, ad := range autodeployers {
		fmt.Printf("%s:%d %s (broken=%v,available=%v)\n", ad.IP, ad.Port, ad.Group, ad.Broken, ad.Available)
		for _, app := range ad.Apps {
			dep := app.Deployment
			fmt.Printf("    %s %d/%s/%s #%d\n", dep.Binary, dep.RepositoryID, dep.Namespace, dep.Groupname, dep.BuildID)
		}
	}
}

func incFailure(ip string, port int32) {
	if *debug {
		fmt.Printf(" Errorcounter increased on %s:%d\n", ip, port)
	}
	ad := getDeployerByAddress(ip, port)
	if ad == nil {
		return
	}
	ad.queryFailures++
}

func getDeployerByAddress(ip string, port int32) *AutoDeployer {
	for _, ad := range autodeployers {
		if (ad.IP == ip) && (ad.Port == port) {
			return ad
		}
	}
	return nil
}

// update an Autodeployer struct
// match on ip/port
func updateAutoDeployer(na *AutoDeployer) {
	for _, ad := range autodeployers {
		if (ad.IP != na.IP) || (ad.Port != na.Port) {
			continue
		}
		ad.Apps = na.Apps
		ad = na
		return
	}
	autodeployers = append(autodeployers, na)
}

// excluding broken ones!
func GetAllDeployersFromCache() []*AutoDeployer {
	var res []*AutoDeployer
	for _, ad := range autodeployers {
		if ad.Broken || (!ad.Available) {
			continue
		}
		res = append(res, ad)
	}
	return res
}

// excluding broken ones!
func GetDeployersFromCache() []*rpb.ServiceAddress {
	var res []*rpb.ServiceAddress
	for _, ad := range autodeployers {
		if ad.Broken {
			continue
		}
		res = append(res, &rpb.ServiceAddress{Host: ad.IP, Port: ad.Port})
	}
	return res
}

// excluding deployments on broken autodeployers
func Cache_GetDeployments(sa *rpb.ServiceAddress, deplid string) []*ad.DeployedApp {
	var res []*ad.DeployedApp
	for _, ad := range autodeployers {
		if (ad.IP != sa.Host) || (ad.Port != sa.Port) || (ad.Broken) {
			continue
		}
		for _, app := range ad.Apps {
			dr := app.Deployment
			if !strings.HasPrefix(dr.DeploymentID, deplid) {
				continue
			}
			res = append(res, app)
		}
	}
	return res
}
