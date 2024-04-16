package deployq

import (
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	"golang.conradwood.net/apis/common"
	"sync"
	"time"
)

const (
	GRACE_PERIOD_BEFORE_SHUT_DOWN = time.Duration(90) * time.Second
)

/*
the second part. stuff has been deployed, now wait for it to perform well before shutting down older versions
it picks up all transactions with the 'started' flag set
*/
func (q *DeployQueue) work_monitoring() {
	t := time.Duration(5) * time.Second
	for {
		time.Sleep(t)

		// find the transactions (with lock held)
		var transactions []*deployTransaction
		var failed_transactions []*deployTransaction
		var new_transactions []*deployTransaction
		q.Lock()
		for _, t := range q.requests {
			if t.err != nil {
				// skipping transactions that failed already
				failed_transactions = append(failed_transactions, t)
				continue
			}
			if t.deployment_processed {
				fmt.Printf("%s processed. removing from queue\n", t.String())
				continue
			}
			new_transactions = append(new_transactions, t)

			if t.started {
				transactions = append(transactions, t)
			}
		}
		q.requests = new_transactions //remove failed transactions from queue
		q.Unlock()

		// deal with the transactions (without lock held)
		for _, t := range transactions {
			err := q.check_monitored(t)
			if err != nil {
				fmt.Printf("monitoring failed: %s\n", err)
			}
		}
		// deal with the failed transactions (without lock held)
		for _, dt := range failed_transactions {
			fmt.Printf("Failed: %s\n", dt.String())
			// stop the ones that were deployed already
			for _, dd := range dt.deployed_ids {
				err := stop_app(dd.Deployer(), dd.ID)
				if err != nil {
					fmt.Printf("monitoring failed: %s\n", err)
				}
			}
		}
	}
}
func (q *DeployQueue) check_monitored(dt *deployTransaction) error {
	var latest_ready time.Time
	all_ready := true
	for _, did := range dt.deployed_ids {
		app := did.deployer.AppByID(did.ID)
		if app == nil {
			if did.running {
				// it was running, but stopped running
				dt.SetError(fmt.Errorf("new version failed unexpectedly on deployer %s", did.deployer.String()))
				dt.sendUpdate(EVENT_FINISHED)
				return nil
			}
			// it has never been running yet. (autodeployer.Deploy() is async!)
			if time.Since(dt.started_time) > time.Duration(60)*time.Second {
				dt.SetError(fmt.Errorf("new version failed to start on deployer %s", did.deployer.String()))
				dt.sendUpdate(EVENT_FINISHED)
				return nil
			}
			all_ready = false
			continue
		}
		did.running = true // it is running, mark as such

		if app.Status == ad.DeploymentStatus_EXECUSER && app.Health == common.Health_READY {
			if !did.ready {
				did.ready = true
				did.ready_time = time.Now()
			}
		}
		if did.ready && did.ready_time.After(latest_ready) {
			latest_ready = did.ready_time
		}
		if !did.ready {
			all_ready = false
		}
	}
	if !all_ready {
		// TODO: timeout here? (allow for 5 minutes or 10 minutes or something before assuming the instance
		// will never become ready?)
		fmt.Printf("DT %s has at least one instance that is not ready yet - no action yet\n", dt.String())
		return nil
	}
	ago := time.Since(latest_ready)
	if ago < GRACE_PERIOD_BEFORE_SHUT_DOWN {
		fmt.Printf("DT %s became ready %0.1f seconds ago - no action yet\n", dt.String(), ago.Seconds())
		return nil
	}
	// TODO: proceed to next step, shutting down previous instances
	fmt.Printf("DT %s is now good (ready since %0.1f seconds ago)\n", dt.String(), ago.Seconds())
	// below just needs to be async
	go completion_stopall(dt)
	return nil
}
func completion_stopall(dt *deployTransaction) {
	if dt.stopping_these {
		return
	}
	dt.stopping_these = true
	fmt.Printf("DT %s: running %d stop requests\n", dt.String(), len(dt.stop_these))
	var err error
	stopping_group := &sync.WaitGroup{}
	for _, dt_stop := range dt.stop_these {
		stopping_group.Add(1)
		go func(dts *deployTransaction_StopRequest) {
			defer stopping_group.Done()
			xerr := stop_app(dts.deployer, dts.deplapp.ID)
			if xerr != nil {
				err = xerr
			} else {
				fmt.Printf("%s Stopped %s on deployer \"%s\"\n", dt.String(), dts.deployer.Host(), dts.deplapp.ID)
			}
		}(dt_stop)
	}
	stopping_group.Wait()
	if err != nil {
		fmt.Printf("%s failed to stop app: %s\n", dt.String(), err)
	}
	dt.deployment_processed = true
}
