package deployq

import (
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/common"
	dp "golang.conradwood.net/deploymonkey/deployplacements"
	"golang.conradwood.net/go-easyops/authremote"
	"strings"
	"sync"
	"time"
)

var (
	// if deployed with "per instance", +1 will be added to the score
	bin_score_match = map[string]int{
		"secureargs-server":  20,
		"logservice-server":  10,
		"errorlogger-server": 10,
		"objectauth-server":  10,
		"objectstore-server": 10,
	}
)

type deployTransaction struct {
	scheduled    bool // true if it is being sent to the worker for processing
	requests     []*dp.DeployRequest
	err          error // set on failure
	result_chan  chan *DeployUpdate
	deployed_ids []*deployed // list of successful deployments
}

func (dt *deployTransaction) Close() {
	close(dt.result_chan)
}
func (dt *deployTransaction) Score() int {
	has_instances := false
	app_score := 0
	for _, r := range dt.requests {
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
	for _, r := range dt.requests {
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
func (dt *deployTransaction) SetError(err error) {
	dt.err = err
	// TODO: send on a channel to notify listeeners
	fmt.Printf("error on deployment: %s\n", err)
}

// caches it on every autodeployer. returns when done
// note: if multiple deployrequests target the same autodeployer it *will* process those concurrently. that may or may not be desired. tbd
func (dt *deployTransaction) CacheEverywhere() error {
	wg := &sync.WaitGroup{}
	var xerr error
	for _, req := range dt.requests {
		wg.Add(1)
		go func(r *dp.DeployRequest) {
			defer wg.Done()
			fmt.Printf("Caching %s on %s\n", r.URL(), r.AutodeployerHost())
			ctx := authremote.ContextWithTimeout(time.Duration(60) * time.Second)
			cl, err := r.GetAutodeployerClient()
			if err != nil {
				xerr = fmt.Errorf("(caching %s): failed to connect to %s: %s", r.URL(), r.AutodeployerHost(), err)
				return
			}
			_, err = cl.CacheURL(ctx, &ad.URLRequest{URL: r.URL()})
			if err != nil {
				xerr = fmt.Errorf("(caching %s): failed to cache on %s: %s", r.URL(), r.AutodeployerHost(), err)
				return
			}
			fmt.Printf("Cached %s on %s\n", r.URL(), r.AutodeployerHost())
		}(req)
	}
	wg.Wait()
	return xerr
}

type deployed struct {
	req *dp.DeployRequest
	ID  string
}

// assuming it is cached everywhere, this will start the appdef
func (dt *deployTransaction) StartEverywhere() error {
	wg := &sync.WaitGroup{}
	var depl_lock sync.Mutex
	var xerr error
	for _, req := range dt.requests {
		wg.Add(1)
		go func(r *dp.DeployRequest) {
			defer wg.Done()
			fmt.Printf("Deploying %s on %s\n", r.URL(), r.AutodeployerHost())
			ctx := authremote.ContextWithTimeout(time.Duration(20) * time.Second)
			cl, err := r.GetAutodeployerClient()
			if err != nil {
				xerr = fmt.Errorf("(deploying %s): failed to connect to %s: %s", r.URL(), r.AutodeployerHost(), err)
				return
			}
			dreq := common.CreateDeployRequest(nil, r.AppDef())
			dr, err := cl.Deploy(ctx, dreq)
			if err != nil {
				xerr = fmt.Errorf("(deploying %s): failed to cache on %s: %s", r.URL(), r.AutodeployerHost(), err)
				return
			}
			depl_lock.Lock()
			dt.deployed_ids = append(dt.deployed_ids, &deployed{req: r, ID: dr.ID})
			depl_lock.Unlock()
			fmt.Printf("deployed %s on %s (ID=%s)\n", r.URL(), r.AutodeployerHost(), dr.ID)
		}(req)
	}
	wg.Wait()

	if xerr != nil {
		// got failure, cleanup all those which were deployed already. Best-effort, ignoring errors
		for _, depl := range dt.deployed_ids {
			ctx := authremote.ContextWithTimeout(time.Duration(20) * time.Second)
			cl, err := depl.req.GetAutodeployerClient()
			if err != nil {
				fmt.Printf("failed to get client to undeploy: %s\n", err)
				continue
			}
			_, err = cl.Undeploy(ctx, &ad.UndeployRequest{ID: depl.ID})
			if err != nil {
				fmt.Printf("failed to undeploy: %s\n", err)
				continue
			}
		}
	}

	return xerr

}
