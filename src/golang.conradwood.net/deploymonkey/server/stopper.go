package main

// responsible for shutting down services which are no longer needed
// (rather: marked for shutdown)
import (
	"flag"
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	rg "golang.conradwood.net/apis/registry"
	dc "golang.conradwood.net/deploymonkey/common"
	"golang.conradwood.net/go-easyops/authremote"
	"sync"
	"time"
)

var (
	debug_cond_stopper = flag.Bool("debug_cond_stopper", false, "debug the conditional stopper code")
	stopDelay          = flag.Duration("stop_delay", time.Duration(0), "delay before shutting down services (after starting new ones)")
	waitExeDelay       = flag.Duration("wait_exe_delay", time.Duration(10)*time.Minute, "delay before giving up on service to become ready")
	filterDelay        = flag.Duration("filter_delay", time.Duration(0), "delay before instructing the registry to filter instances marked for shutdown")
	stopRequests       []*stopRequest
	stoptransactionctr int
	translock          sync.Mutex
	reqlock            sync.Mutex
	regClient          rg.RegistryClient
)

func init() {
	go stopperThread()
}

type StopperCondition interface {
	// 0=n/a, 1=inconclusive, 2=FALSE, 3=TRUE
	// on 0/1 no change to the stoprequest
	// if one ore more conditions are 2 : transaction abort!
	// if all conditions are 3 : start filtering immediately
	eval(s *stopRequest) (int, error)
	String() string
}

type stopRequest struct {
	transaction    int
	prefix         string
	submitted      time.Time
	waitUntilReady time.Time // not yet used
	executeAt      time.Time
	filterAt       time.Time
	id             string
	autodeployer   *rg.ServiceAddress
	ports          []uint32 // ports the instance is listening on
	done           bool
	cancelled      bool
	failurectr     int
	filtered       bool
	deployInfo     *ad.DeployInfo
	conditions     []StopperCondition
	lastEval       int // last result from conditions
}

func (s *stopRequest) AddCondition(sc StopperCondition) {
	s.conditions = append(s.conditions, sc)
}
func (s *stopRequest) String() string {
	ps := ""
	for _, p := range s.ports {
		ps = ps + fmt.Sprintf("%d ", p)
	}
	return fmt.Sprintf("%s:%s", s.autodeployer.Host, ps)
}

func stopperThread() {
	for {
		if *debug {
			fmt.Printf("stopper/filterer: Got %d requests...\n", len(stopRequests))
		}
		cleanRequestQueue()
		// Pattern:
		// 1. lock()
		// 2. build list of stuff
		// 3. unlock()
		// 4. execute list
		// why? the cancel() and add() functions should not
		// block for longer than necessary
		// race conditions acceptable
		// (we're using long running time-based processes)

		/************************** conditions *************************/

		condRun := stopRequests
		var user_messages []string
		for _, f := range condRun {
			res, err := conditionExecute(f)
			if err != nil {
				fmt.Printf("failed to eval condition %s: %s\n", f.prefix, err)
			} else {
				s := fmt.Sprintf("Result of condition \"%s\": %d\n", f.String(), res)
				user_messages = append(user_messages, s)
				f.lastEval = res
				if f.lastEval == 2 {
					cancelStop(f.transaction, user_messages, "Deployment did not succeed fully. Current Version NOT undeployed.")
				}
			}
		}

		cleanRequestQueue()

		/************************** filtering *************************/

		// pick the ones we want to filter
		var filter []*stopRequest
		reqlock.Lock()
		// note how we are not checking 'req.filtered'
		// we tell the registry to filter for 10 seconds
		// and repeat every 5s.
		// thus, if deploymonkey crashes or weird stuff happens
		// the previous instances will become available
		// again after 10s max
		for _, req := range stopRequests {
			if (req.lastEval != 3) && (req.filterAt.After(time.Now())) {
				continue
			}
			filter = append(filter, req)
		}
		reqlock.Unlock()

		// filter the ones we picked
		for _, f := range filter {
			err := filterExecute(f)
			if err != nil {
				fmt.Printf("failed to filter %s: %s\n", f.prefix, err)
			} else {
				f.filtered = true
			}
		}

		cleanRequestQueue()

		/************************** execution *************************/

		// pick the ones we want to execute
		var exe []*stopRequest
		reqlock.Lock()
		for _, req := range stopRequests {
			if req.executeAt.After(time.Now()) {
				continue
			}
			if !req.filtered {
				fmt.Printf("Not stopping %s - it is not filtered yet\n", req.String())
				continue
			}
			exe = append(exe, req)
		}
		reqlock.Unlock()

		// execute the ones we picked
		for _, e := range exe {
			err := stopExecute(e)
			if err != nil {
				fmt.Printf("failed to stop %s: %s\n", e.prefix, err)
			} else {
				e.done = true
			}
		}
		cleanRequestQueue()
		time.Sleep(time.Duration(5) * time.Second)
	}
}

// remove the ones which are processed
func cleanRequestQueue() {
	var ns []*stopRequest
	reqlock.Lock()
	tooOld := time.Now().Add(time.Duration(-30) * time.Minute)
	for _, req := range stopRequests {
		if req.submitted.Before(tooOld) {
			fmt.Printf("Dropped %s - too old (%v)\n", req.String(), req.submitted)
			continue
		}
		if req.done || req.cancelled {
			continue
		}
		if req.lastEval == 2 {
			continue
		}
		ns = append(ns, req)
	}
	stopRequests = ns
	reqlock.Unlock()

}

// filter an instance (from registry)
func filterExecute(sr *stopRequest) error {
	//	fmt.Printf("Filter execute disabled. Hidden service should be unnecessary\n")
	return nil
}

// stop an instance NOW
func stopExecute(sr *stopRequest) error {
	var err error
	s := ""
	if sr.deployInfo != nil {
		s = sr.deployInfo.Binary
	}
	fmt.Printf("Shutting down: %s (%s) on %s\n", sr.id, s, sr.autodeployer.Host)
	conn, err := DialService(sr.autodeployer)
	if err != nil {
		return fmt.Errorf("Failed to connect to autodeployer %v", sr.autodeployer)
	}
	defer conn.Close()
	adc := ad.NewAutoDeployerClient(conn)

	ud := ad.UndeployRequest{ID: sr.id}
	ures, err := adc.Undeploy(authremote.Context(), &ud)
	if err != nil {
		fmt.Printf("Failed to shutdown %s @ %s: %s\n", sr.id, sr.autodeployer.Host, err)
		return fmt.Errorf("Failed to shutdown %s @ %s: %s\n", sr.id, sr.autodeployer.Host, err)
	}
	fmt.Printf("Undeployed request sent, confirmed ID %s\n", ures.ID)
	return nil
}

// returns a transaction number to refer to all generated stop requests
func stop(stopPrefix string) (int, error) {
	trans := stopTransaction()
	fmt.Printf("Stopping prefix %s\n", stopPrefix)
	sas, err := GetDeployers()
	if err != nil {
		return 0, err
	}
	fmt.Printf("Looking for services to stop with deployment id prefix: \"%s\"\n", stopPrefix)
	// this is way... to dumb. we do two steps:
	// 1. shutdown all applications for this group
	// 2. fire up new ones
	// TODO: make it smart and SAFELY apply diffs when and if necessary...

	// stopping stuff...
	for _, sa := range sas {
		fmt.Printf("Querying service at: %s:%d\n", sa.Host, sa.Port)
		conn, err := DialService(sa)
		if err != nil {
			fmt.Printf("Failed to connect to service %v\n", sa)
			continue
		}
		adc := ad.NewAutoDeployerClient(conn)

		apps, err := getDeployments(adc, sa, stopPrefix)
		if err != nil {
			fmt.Printf("Queried service (%v)\n", err)
			conn.Close()
			fmt.Printf("Unable to get deployments from %v: %s\n", sa, err)
			continue
		}
		fmt.Printf("Queried service %s:%d\n", sa.Host, sa.Port)

		/*******************************************************
		// set the timers on what happens next
		*****************************************************/
		exeTime := time.Now().Add(*stopDelay)
		waitReadyTime := time.Now().Add(*waitExeDelay)
		filterTime := time.Now().Add(*filterDelay)
		if (*stopDelay != 0) && ((*filterDelay == 0) || (*filterDelay >= *stopDelay)) {
			// filterdelay is 0 or >= stopDelay
			nf := time.Duration(0)
			// if no filter delay specified,
			// default is 45 seconds before stopDelay
			if *filterDelay == time.Duration(0) {
				nf = *stopDelay - (time.Duration(45) * time.Second)
			}
			if nf < 0 {
				nf = time.Duration(5) * time.Second
			}
			if nf >= *stopDelay {
				nf = time.Duration(1) * time.Second
			}
			filterTime = time.Now().Add(nf)
		}

		/********************************************
		// find and process each instance
		********************************************/
		for _, deplApp := range apps {
			deployInfo := deplApp.Deployment
			fmt.Printf("Need to stop: %s (status=%s)\n", deployInfo.Binary, deployInfo.Status)

			sr := &stopRequest{
				transaction:    trans,
				prefix:         stopPrefix,
				submitted:      time.Now(),
				executeAt:      exeTime,
				waitUntilReady: waitReadyTime,
				filterAt:       filterTime,
				autodeployer:   sa,
				id:             deplApp.ID,
				ports:          deployInfo.Ports,
				deployInfo:     deployInfo,
				done:           false,
			}
			if *stopDelay != 0 {
				addStop(sr)
			} else {
				stopExecute(sr)
			}
		}
		conn.Close()
	}
	return trans, nil
}

func stopTransaction() int {
	translock.Lock()
	defer translock.Unlock()
	if stoptransactionctr > 10000 {
		stoptransactionctr = 0
	}
	stoptransactionctr++
	return stoptransactionctr
}

// we changed our minds and do not want to stop
// current versions.
// e.g. we might have had trouble deploying the
// new instances
func cancelStop(transaction int, logmessages []string, errmsg string) {
	reqlock.Lock()
	var sreq *stopRequest
	for _, sr := range stopRequests {
		// we don't cancel "twice" because
		// a) it's silly and b) we don't want nag people via slack
		if (sr.transaction == transaction) && (sr.cancelled == false) {
			sr.cancelled = true
			sreq = sr
		}
	}
	reqlock.Unlock()
	if sreq != nil {
		NotifyPeopleAboutCancel(sreq, logmessages, errmsg)
	}
}

func addStop(sr *stopRequest) {
	fmt.Printf("Queueing request to stop %s\n", sr.String())
	reqlock.Lock()
	defer reqlock.Unlock()
	stopRequests = append(stopRequests, sr)
}

// this attaches a condition to the "stopper".
// specifically the stopper monitors the instance specified by (startupid,startuphost)
// if the instance is running for more than `minruntime` seconds, it will proceed
// to filter previous ones immediately.
// if the instance has stopped, the transaction will be aborted.
// (filters removed and stop cancelled)
func stopperRunningCondition(transaction int, startupid string, startuphost *rg.ServiceAddress, minruntime int) {
	src := &StopperRunningCondition{startupid: startupid, host: startuphost, minruntime: minruntime}
	reqlock.Lock()
	for _, sr := range stopRequests {
		// we don't cancel "twice" because
		// a) it's silly and b) we don't want nag people via slack
		if sr.transaction == transaction {
			sr.AddCondition(src)
		}
	}
	reqlock.Unlock()
}

func GetDeployments(host string) (*ad.InfoResponse, error) {
	port := 4000
	conn, err := DialService(&rg.ServiceAddress{Host: host, Port: int32(port)})
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to service %s:%d", host, port)
	}
	defer conn.Close()
	adc := ad.NewAutoDeployerClient(conn)
	ctx := authremote.Context()
	info, err := adc.GetDeployments(ctx, dc.CreateInfoRequest())
	return info, err
}

// execute all conditions of this stoprequest
// if at least one is 2: return 2
// if all are 3: return 3
// otherwise 1
// if any throw an error - return the error)
func conditionExecute(s *stopRequest) (int, error) {
	if len(s.conditions) == 0 {
		return 0, nil
	}
	for _, c := range s.conditions {
		r, err := c.eval(s)
		if err != nil {
			if *debug_cond_stopper {
				fmt.Printf("Condition \"%s\" failed to execute: %s\n", c.String(), err)
			}
			return 0, err
		}
		if *debug_cond_stopper {
			fmt.Printf("Condition \"%s\" returned %d\n", c.String(), r)
		}
		if r == 2 {
			if *debug_cond_stopper {
				fmt.Printf("Conditions evaluated to FALSE\n")
			}

			return 2, nil
		}
		if r == 3 {
			if *debug_cond_stopper {
				fmt.Printf("Condition \"%s\" preliminary result: TRUE\n", c.String())
			}
			continue
		}
		fmt.Printf("Conditions evaluated to INCONCLUSIVE\n")
		return 1, nil
	}
	if *debug_cond_stopper {
		fmt.Printf("Conditions evaluated to TRUE\n")
	}
	return 3, nil
}
