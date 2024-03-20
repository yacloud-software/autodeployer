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

var (
	q = &DeployQueue{
		autodeployer_locks:    make(map[string]bool),
		work_distributor_chan: make(chan bool),
		work_handler_chan:     make(chan *deployTransaction),
	}
	starterlock sync.Mutex
)

// add a bunch of requests, treat them somewhat as one transaction
func Add(dr []*dp.DeployRequest) {

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
	tr := &deployTransaction{requests: dr}
	q.requests = append(q.requests, tr)
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
		q.Lock()
		var next *deployTransaction
		for _, dt := range q.requests {
			if q.hasLockedAutodeployers(dt) {
				continue
			}
			if next == nil || dt.Score() > next.Score() {
				next = dt
			}
		}
		q.Unlock()
		if next != nil {
			q.work_handler_chan <- next
		}
	}
}

func (q *DeployQueue) hasLockedAutodeployers(dt *deployTransaction) bool {
	for _, r := range dt.requests {
		host := r.AutodeployerHost()
		if q.autodeployer_locks[host] {
			return true
		}
	}
	return false
}

func (q *DeployQueue) work_handler() {
	for {
		dt := <-q.work_handler_chan
		fmt.Printf("work handling: %#v\n", dt)
	}
}
