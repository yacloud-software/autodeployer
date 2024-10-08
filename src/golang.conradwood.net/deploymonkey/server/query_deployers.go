package main

import (
	"flag"
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	dm "golang.conradwood.net/apis/deploymonkey"
	rpb "golang.conradwood.net/apis/registry"
	dc "golang.conradwood.net/deploymonkey/common"
	"golang.conradwood.net/deploymonkey/config"
	"golang.conradwood.net/deploymonkey/scheduler"
	"golang.conradwood.net/deploymonkey/suggest"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/prometheus"
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
	timediff_deployers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "deploymonkey_deployer_timediff",
			Help: "V=1 UNIT=seconds DESC=time difference between deploymonkey and autodeployer",
		},
		[]string{"adinstance"},
	)
	lastscan = &scanner{}
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
type scanner struct {
	sync.Mutex
	counter     int
	deployments []*ad.InfoResponse
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
	prometheus.MustRegister(timediff_deployers, suggestioncount, suggestion_missing_deployers, machinegroup_deployers)
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
			dl, err := INT_GetDeploymentsFromCache(authremote.Context())
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
	wg := &sync.WaitGroup{}
	sc := &scanner{}
	for _, xsa := range sas {
		wg.Add(1)
		go func(sa *rpb.ServiceAddress) {
			defer wg.Done()
			err = sc.ScanAutodeployer(sa)
			if err != nil {
				fmt.Printf("Error scanning %s: %s\n", sa.Host, err)
			}
		}(xsa)
	}
	wg.Wait()
	for _, sa := range sas {
		setBrokennessOnDeployers(sa)
	}
	lastscan = sc
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

func (sc *scanner) ScanAutodeployer(sa *rpb.ServiceAddress) error {
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
		return errors.Errorf("%s", s)
	}
	adc := ad.NewAutoDeployerClient(conn)
	req := &ad.MachineInfoRequest{}
	ctx := authremote.Context()
	mir, err := adc.GetMachineInfo(ctx, req)
	now := time.Now()
	if err != nil {
		s := fmt.Sprintf("Failed to get machineinfo on %s: %s", sa.Host, err)
		fmt.Println(s)
		incFailure(sa.Host, sa.Port)
		conn.Close()
		return errors.Errorf("%s", s)
	}
	ad_time := time.Unix(int64(mir.CurrentTime), 0)
	diff := now.Sub(ad_time)
	timediff_deployers.With(prometheus.Labels{"adinstance": sa.Host}).Set(diff.Seconds())
	da, err := adc.GetDeployments(ctx, dc.CreateInfoRequest())
	if err != nil {
		incFailure(sa.Host, sa.Port)
		conn.Close()
		return errors.Errorf("Failed to get deployment info on %s: %s\n", sa.Host, err)
	}
	conn.Close()
	nt.queryFailures = 0
	nt.Group = mir.MachineGroup
	nt.LastSeen = time.Now()
	nt.Apps = da.Apps
	if *debug {
		fmt.Printf("Found %d apps on %s:%d\n", len(nt.Apps), sa.Host, sa.Port)
	}
	sc.AddDeployments(da)
	updateAutoDeployer(nt)
	return nil
}
func (sc *scanner) AddDeployments(da *ad.InfoResponse) {
	sc.Lock()
	sc.deployments = append(sc.deployments, da)
	sc.Unlock()
}
func GetLastQueryResult() *scanner {
	return lastscan
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
