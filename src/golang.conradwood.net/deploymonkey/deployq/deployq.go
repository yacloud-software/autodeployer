/*
		efficiently handle deployrequests. that is, maintain a queue and pick the next one to deal with.
		locking by autodeployer instance and application definition.
	        This isn't quite a queue, because subsequent deployrequests might cancel out previous ones.
	        For example: "deploy application foo in version 5 on 3 autodeployers", subsequent submission of
	        "deploy application foo in version 6 on 3 autodeployers cancels the first one out".
*/
package deployq

import (
	"fmt"
	dp "golang.conradwood.net/deploymonkey/deployplacements"
	"sync"
)

type EVENT int

const (
	EVENT_CACHE    = 1
	EVENT_START    = 2
	EVENT_PREPARE  = 3
	EVENT_ERROR    = 4
	EVENT_FINISHED = 5
)

var (
	q = &DeployQueue{
		autodeployer_locks:    make(map[string]bool),
		work_distributor_chan: make(chan bool),
		work_handler_chan:     make(chan *deployTransaction),
	}
	starterlock sync.Mutex
)

// add a bunch of requests, treat them somewhat as one transaction
func Add(dr []*dp.DeployRequest) chan *DeployUpdate {

	// start worker if necessary
	starterlock.Lock()
	if !q.workers_started {
		go q.work_distributor()
		go q.work_handler()
		q.workers_started = true
	}
	starterlock.Unlock()

	// add to queue
	q.Lock()
	tr := &deployTransaction{
		requests:    dr,
		result_chan: make(chan *DeployUpdate, 100),
	}
	q.requests = append(q.requests, tr)
	q.work_distributor_chan <- true
	q.Unlock()
	return tr.result_chan
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
		for {
			q.Lock()
			var next *deployTransaction
			for _, dt := range q.requests {
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

// call with q.lock() held
func (q *DeployQueue) hasLockedAutodeployers(dt *deployTransaction) bool {
	for _, r := range dt.requests {
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
			return fmt.Errorf("autodeployer %s locked already", host)
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
func (q *DeployQueue) work_handler() {
	for {
		dt := <-q.work_handler_chan
		fmt.Printf("work handling: %#v\n", dt)
		dt.sendUpdate(EVENT_PREPARE)
		q.Lock()
		err := q.lockTransaction(dt)
		if err != nil {
			dt.SetError(fmt.Errorf("failed to lock transaction (%w)", err))
			q.Unlock()
			dt.sendUpdate(EVENT_FINISHED)
			dt.Close()
			continue
		}
		q.Unlock()

		dt.sendUpdate(EVENT_CACHE)
		// now cache it everywhere
		err = dt.CacheEverywhere()
		if err != nil {
			dt.SetError(err)
			q.unlockTransaction(dt)
			dt.sendUpdate(EVENT_FINISHED)
			dt.Close()
			continue
		}
		dt.sendUpdate(EVENT_START)
		// now start it everywhere
		err = dt.StartEverywhere()
		if err != nil {
			dt.SetError(err)
			q.unlockTransaction(dt)
			dt.sendUpdate(EVENT_FINISHED)
			dt.Close()
			continue
		}
		q.unlockTransaction(dt)
		dt.sendUpdate(EVENT_FINISHED)
		dt.Close()
	}
}
