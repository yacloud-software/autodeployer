/*
		efficiently handle deployrequests. that is, maintain a queue and pick the next one to deal with.
		locking by autodeployer instance and application definition.
	        This isn't quite a queue, because subsequent deployrequests might cancel out previous ones.
	        For example: "deploy application foo in version 5 on 3 autodeployers", subsequent submission of
	        "deploy application foo in version 6 on 3 autodeployers cancels the first one out".

the deployq has multiple independent workers:
1) cache & deploy
2) monitor new deployments for crashes and wait for all to report "ready" status
3) shut down old ones
4) communicate error or success to users (via slack)

each transaction is processed by one of the workers until either it encounters an error (SetError()) or success (SetSuccess())
*/
package deployq

import (
	"context"
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/common"
	"golang.conradwood.net/deploymonkey/db"
	dp "golang.conradwood.net/deploymonkey/deployplacements"
	"golang.conradwood.net/go-easyops/errors"
	"sort"
	"sync"
	"time"
)

type EVENT int

const (
	EVENT_CACHE    = 1
	EVENT_START    = 2
	EVENT_PREPARE  = 3
	EVENT_ERROR    = 4
	EVENT_FINISHED = 5
	EVENT_STARTED  = 6
)

var (
	debug = flag.Bool("debug_deployq", true, "debug the deployq")
	q     = &DeployQueue{
		autodeployer_locks:    make(map[string]bool),
		work_distributor_chan: make(chan bool),
		work_handler_chan:     make(chan *deployTransaction),
	}
	starterlock sync.Mutex
)

// add a bunch of requests, treat them somewhat as one transaction
func Add(dr []*dp.DeployRequest) (chan *DeployUpdate, error) {
	if len(dr) == 0 {
		return nil, errors.Errorf("0 deployrequests received")
	}
	// start worker if necessary
	starterlock.Lock()
	if !q.workers_started {
		go q.work_distributor()
		go q.work_handler()
		go q.work_monitoring()
		q.workers_started = true
	}
	starterlock.Unlock()
	for _, r := range dr {
		if len(r.AppDef().Args) == 0 {
			panic(fmt.Sprintf("no args for %d(%s)", r.AppDef().ID, r.AppDef().Binary))
		}
	}
	// add to queue
	q.Lock()
	tr := &deployTransaction{
		start_requests:             dr,
		result_chan:                make(chan *DeployUpdate, 100),
		stop_running_in_same_group: true,
	}
	debugf("adding deploytransaction %s", tr.String())
	q.requests = append(q.requests, tr)
	q.work_distributor_chan <- true
	q.Unlock()
	return tr.result_chan, nil
}

type DeployUpdate struct {
	event EVENT
	err   error
}

type DeployQueue struct {
	sync.Mutex
	requests              []*deployTransaction
	autodeployer_locks    map[string]bool // host ip, true/false, must be accessed with deployqueue.Lock()
	work_distributor_chan chan bool
	work_handler_chan     chan *deployTransaction
	workers_started       bool
}

func (q *DeployQueue) work_distributor() {
	for {
		<-q.work_distributor_chan

		//save them to database
		ctx := context.Background()
		for _, dt := range q.requests {
			for _, sr := range dt.start_requests {
				dl := &pb.DeploymentLog{
					BuildID:          sr.AppDef().BuildID,
					AppDef:           sr.AppDef(),
					Binary:           sr.AppDef().Binary,
					DeployAlgorithm:  2,
					AutoDeployerHost: sr.Deployer().String(),
					Started:          uint32(time.Now().Unix()),
				}
				db.DefaultDBDeploymentLog().Save(ctx, dl)
				dt.AddLog(dl)
			}
		}

		for {
			q.Lock()
			var next *deployTransaction
			for _, dt := range q.requests {
				debugf("processing deploytransaction %s", dt.String())

				if dt.scheduled {
					continue
				}
				if q.hasLockedAutodeployers(dt) {
					continue
				}
				if next == nil || dt.Score() > next.Score() {
					next = dt
				}
			}
			q.Unlock()
			if next == nil {
				break
			}

			next.scheduled = true
			q.work_handler_chan <- next
		}
	}
}
func (q *DeployQueue) remove(dt *deployTransaction) {
	q.Lock()
	defer q.Unlock()
	var res []*deployTransaction
	for _, r := range q.requests {
		if r == dt {
			continue
		}
		res = append(res, r)
	}
	q.requests = res
}

// call with q.lock() held
func (q *DeployQueue) hasLockedAutodeployers(dt *deployTransaction) bool {
	for _, r := range dt.start_requests {
		host := r.AutodeployerHost()
		if q.autodeployer_locks[host] {
			return true
		}
	}
	return false
}

// call with q.lock() held
func (q *DeployQueue) lockAutodeployers(dt *deployTransaction) error {
	for _, host := range dt.AutodeployerHosts() {
		b := q.autodeployer_locks[host]
		if b {
			return errors.Errorf("autodeployer %s locked already", host)
		}
		q.autodeployer_locks[host] = true

	}
	return nil
}

// call with q.lock() held
func (q *DeployQueue) unlockAutodeployers(dt *deployTransaction) {
	for _, host := range dt.AutodeployerHosts() {
		q.autodeployer_locks[host] = false

	}
}

// call with q.lock() held
func (q *DeployQueue) lockApplications(dt *deployTransaction) error {
	// this is a noop at the moment
	return nil
}

// call with q.lock() held
func (q *DeployQueue) unlockApplications(dt *deployTransaction) {
	// this is a noop at the moment
}

// call with q.lock() held
func (q *DeployQueue) lockTransaction(dt *deployTransaction) error {
	err := q.lockAutodeployers(dt)
	if err != nil {
		return err
	}
	err = q.lockApplications(dt)
	if err != nil {
		return err
	}

	return nil
}

// call WITHOUT q.lock() held
func (q *DeployQueue) unlockTransaction(dt *deployTransaction) {
	q.Lock()
	defer q.Unlock()
	q.unlockApplications(dt)
	q.unlockAutodeployers(dt)
}

/*
this is part 1 of the deployment process. this is where the application is cached and started on each autodeployer that is relevant.
if something goes wrong, it will set an error on the transaction.
if all goes well it sets the 'started' flag on the transaction. (which then gets picked up by the completion worker)
*/
func (q *DeployQueue) work_handler() {
	for {
		dt := <-q.work_handler_chan
		fmt.Printf("work handling: %#v\n", dt)
		dt.sendUpdate(EVENT_PREPARE)
		q.Lock()
		err := q.lockTransaction(dt)
		if err != nil {
			dt.SetError(errors.Errorf("failed to lock transaction (%w)", err))
			q.Unlock()
			dt.sendUpdate(EVENT_FINISHED)
			continue
		}
		q.Unlock()

		// create deploymentids for each request
		for _, dr := range dt.start_requests {
			appdef := dr.AppDef()
			if appdef.DeploymentID == "" {
				appdef.DeploymentID = common.CreateDeploymentID(appdef)

			}
		}

		// now cache it everywhere
		err = dt.CacheEverywhere()
		if err != nil {
			dt.SetError(err)
			q.unlockTransaction(dt)
			dt.sendUpdate(EVENT_FINISHED)
			continue
		}
		dt.sendUpdate(EVENT_START)

		if dt.stop_running_in_same_group {
			dt.find_stoppers()
		}

		// now start it everywhere
		err = dt.StartEverywhere()
		if err != nil {
			dt.SetError(err)
			q.unlockTransaction(dt)
			dt.sendUpdate(EVENT_FINISHED)
			continue
		}
		dt.sendUpdate(EVENT_STARTED)
		q.unlockTransaction(dt)
		dt.started = true // means it will now be monitored
		dt.started_time = time.Now()
	}
}

func debugf(format string, args ...interface{}) {
	if *debug {
		return
	}
	s := fmt.Sprintf(format, args...)
	fmt.Printf("[deployq] %s\n", s)
}

// find apps that need stopping if this transaction is successful
// stores them in dt_stop_these variable
func (dt *deployTransaction) find_stoppers() {
	// before we start it, get a list of applications that we will need to stop if this deployment is successful
	var stop_apps []*deployTransaction_StopRequest
	for _, dr := range dt.start_requests {
		da := common.FindByAppDef(dr.AppDef())
		for _, ade := range da {
			found := false
			for _, st := range stop_apps {
				if st.deployer.Host() != ade.Deployer().Host() {
					continue
				}
				if st.deplapp.ID != ade.DeployedApp().ID {
					continue
				}
				found = true
				break
			}
			if found {
				continue
			}
			dst := &deployTransaction_StopRequest{
				deployer: ade.Deployer(),
				deplapp:  ade.DeployedApp(),
			}
			stop_apps = append(stop_apps, dst)
		}
	}
	dt.stop_these = stop_apps
	fmt.Printf("%s - if transaction is succesful, stop %d prior apps\n", dt.String(), len(dt.stop_these))
	sort.Slice(dt.stop_these, func(i, j int) bool {
		return dt.stop_these[i].deplapp.DeployRequest.DeploymentID < dt.stop_these[j].deplapp.DeployRequest.DeploymentID
	})
	for _, st := range dt.stop_these {
		fmt.Printf("   %s: %s %s\n", dt.String(), st.deplapp.DeployRequest.DeploymentID, st.deplapp.DeployRequest.AppReference.AppDef.Binary)
	}

}
