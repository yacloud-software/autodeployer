package deployq

import (
	"context"
	"flag"
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/common"
	"golang.conradwood.net/deploymonkey/db"
	dp "golang.conradwood.net/deploymonkey/deployplacements"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/errors"
	"strings"
	"sync"
	"time"
)

var (
	deploy_timeout = flag.Duration("deploy_timeout", time.Duration(40)*time.Second, "timeout after which to cancel a deployrequest")
	// if deployed with "per instance", +1 will be added to the score
	bin_score_match = map[string]int{
		"secureargs-server":  20,
		"logservice-server":  10,
		"errorlogger-server": 10,
		"objectauth-server":  10,
		"objectstore-server": 10,
	}
)

type deployTransaction_StopRequest struct {
	deployer       *common.Deployer
	deplapp        *ad.DeployedApp
	stopping       bool
	stopping_since time.Time
	stopped        bool
}

func (dts *deployTransaction_StopRequest) String() string {
	return fmt.Sprintf("%s %s on %s", dts.deplapp.ID, dts.deplapp.DeployRequest.Binary, dts.deployer.String())
}

type deployTransaction struct {
	scheduled                  bool // true if it is being sent to the worker for processing
	start_requests             []*dp.DeployRequest
	err                        error // set on failure
	result_chan                chan *DeployUpdate
	deployed_ids               []*deployed // list of successful deployments
	stop_running_in_same_group bool        // if true, after successful deployment, undeploy versions in the same group other than the ones just started
	started                    bool        // true if stuff has been successfully started (and is expected to now be monitored until older versions can be shut down)
	started_time               time.Time
	stop_these                 []*deployTransaction_StopRequest
	stopping_these             bool // true if stop prior apps is already in progress
	deployment_processed       bool // if true, nothing further to do
	logs                       []*pb.DeploymentLog
}

func (dt *deployTransaction) AddLog(dl *pb.DeploymentLog) {
	dt.logs = append(dt.logs, dl)
}
func (dt *deployTransaction) String() string {
	x := ""
	v := uint64(0)
	if len(dt.start_requests) > 0 {
		x = dt.start_requests[0].AppDef().Binary
		v = dt.start_requests[0].AppDef().BuildID
	}
	return fmt.Sprintf("deploytransaction %d deployrequests, first binary: \"%s\" (Version=%d)", len(dt.start_requests), x, v)
}

func (dt *deployTransaction) Close() {
	close(dt.result_chan)
}
func (dt *deployTransaction) Score() int {
	has_instances := false
	app_score := 0
	for _, r := range dt.start_requests {
		appdef := r.AppDef()
		if appdef.InstancesMeansPerAutodeployer {
			has_instances = true
		}
		as := appScore(appdef)
		if as > app_score {
			app_score = as
		}
	}
	res := app_score
	if has_instances {
		res++
	}
	return res
}
func (dt *deployTransaction) AutodeployerHosts() []string {
	rm := make(map[string]bool)
	for _, r := range dt.start_requests {
		rm[r.AutodeployerHost()] = true
	}
	var res []string
	for k, _ := range rm {
		res = append(res, k)
	}
	return res
}

func appScore(ad *pb.ApplicationDefinition) int {
	for k, v := range bin_score_match {
		if strings.Contains(ad.Binary, k) {
			return v
		}
	}
	return 0
}
func (dt *deployTransaction) SetSuccess() {
	fmt.Printf("Transaction %s completed successfully\n", dt.String())
	dt.Close()
	//TODO: do something here, like tell the user
}
func (dt *deployTransaction) SetError(err error) {
	dt.deploylogs_set_error(err)
	dt.err = err
	// TODO: send on a channel to notify listeners
	fmt.Printf("error on deployment: %s\n", errors.ErrorStringWithStackTrace(err))
}

// caches it on every autodeployer. returns when done
// note: if multiple deployrequests target the same autodeployer it *will* process those concurrently. that may or may not be desired. tbd
func (dt *deployTransaction) CacheEverywhere() error {
	wg := &sync.WaitGroup{}
	var xerr error
	for _, req := range dt.start_requests {
		wg.Add(1)
		go func(r *dp.DeployRequest) {
			defer wg.Done()
			fmt.Printf("Caching %s on %s\n", r.DownloadURL(), r.AutodeployerHost())
			ctx := authremote.ContextWithTimeout(time.Duration(60) * time.Second)
			cl := r.GetAutodeployerClient()
			_, err := cl.CacheURL(ctx, &ad.URLRequest{URL: r.DownloadURL()})
			if err != nil {
				xerr = errors.Errorf("(caching %s): failed to cache on %s: %s", r.DownloadURL(), r.AutodeployerHost(), err)
				return
			}
			fmt.Printf("Cached %s on %s\n", r.DownloadURL(), r.AutodeployerHost())
		}(req)
	}
	wg.Wait()
	return xerr
}

type deployed struct {
	req        *dp.DeployRequest
	ID         string // the autodeployer ID
	deployer   *common.Deployer
	ready_time time.Time
	ready      bool
	running    bool // true if autodeployer reports this at least once
}

func (dd *deployed) Deployer() *common.Deployer {
	return dd.deployer
}

// assuming it is cached everywhere, this will start the appdef
func (dt *deployTransaction) StartEverywhere() error {
	wg := &sync.WaitGroup{}
	var depl_lock sync.Mutex
	var xerr error
	for _, req := range dt.start_requests {
		wg.Add(1)
		go func(r *dp.DeployRequest) {
			defer wg.Done()
			fmt.Printf("Deploying %s on %s\n", r.URL(), r.AutodeployerHost())
			started := time.Now()
			ctx := authremote.ContextWithTimeout(*deploy_timeout)
			cl := r.GetAutodeployerClient()
			dreq := common.CreateDeployRequest(nil, r.AppDef())
			dr, err := cl.Deploy(ctx, dreq)
			dur := time.Since(started)
			if err != nil {
				xerr = errors.Errorf("(deploying %s): failed to deploy on %s: %s", r.String(), r.AutodeployerHost(), err)
				return
			}
			depl_lock.Lock()
			dd := &deployed{deployer: r.Deployer(), req: r, ID: dr.ID}
			dt.deployed_ids = append(dt.deployed_ids, dd)
			depl_lock.Unlock()
			fmt.Printf("deployed %s on %s (ID=%s), took %0.1fs\n", r.String(), r.AutodeployerHost(), dr.ID, dur.Seconds())
		}(req)
	}
	wg.Wait()

	if xerr != nil {
		dt.deploylogs_set_error(xerr)
		// got failure, cleanup all those which were deployed already. Best-effort, ignoring errors
		for _, depl := range dt.deployed_ids {
			ctx := authremote.ContextWithTimeout(time.Duration(20) * time.Second)
			cl := depl.req.GetAutodeployerClient()
			_, err := cl.Undeploy(ctx, &ad.UndeployRequest{ID: depl.ID})
			if err != nil {
				fmt.Printf("failed to undeploy: %s\n", err)
				continue
			}
		}
	}

	return xerr

}

func (dt *deployTransaction) sendUpdate(ev EVENT) {
	du := &DeployUpdate{
		event: ev,
		err:   dt.err,
	}
	dt.result_chan <- du
}

func (dt *deployTransaction) deploylogs_set_error(err error) {
	ctx := context.Background()
	for _, dl := range dt.logs {
		dl.Message = fmt.Sprintf("%s", err)
		db.DefaultDBDeploymentLog().Update(ctx, dl)
	}
}
