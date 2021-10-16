package changes

import (
	pb "golang.conradwood.net/apis/deployminator"
	"sync"
)

var (
	olock    sync.Mutex
	locks_dd = make(map[uint64]*sync.Mutex)
	locks_ad = make(map[string]*sync.Mutex)
)

type Lock interface {
	Unlock()
}

type lock struct {
	dd   *pb.DeploymentDescriptor
	ad   string
	lock *sync.Mutex
}

func Lock_DeploymentDescriptor(dd *pb.DeploymentDescriptor) Lock {
	olock.Lock()
	ld := locks_dd[dd.ID]
	if ld == nil {
		ld = &sync.Mutex{}
		locks_dd[dd.ID] = ld
	}
	olock.Unlock()
	ld.Lock()
	return &lock{lock: ld, dd: dd}
}

func Lock_Autodeployer(address string) Lock {
	olock.Lock()
	ld := locks_ad[address]
	if ld == nil {
		ld = &sync.Mutex{}
		locks_ad[address] = ld
	}
	olock.Unlock()
	ld.Lock()
	return &lock{lock: ld, ad: address}
}

func (l *lock) Unlock() {
	l.lock.Unlock()
}
