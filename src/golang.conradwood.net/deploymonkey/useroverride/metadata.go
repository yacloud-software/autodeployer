package useroverride

import (
	"sync"

	pb "golang.conradwood.net/apis/deploymonkey"
)

var (
	md_lock  sync.Mutex
	metadata []*MetaData
)

type MetaData struct {
	appdef              *pb.ApplicationDefinition
	user_requested_stop bool
}

func GetOrCreateMetaData(appdef *pb.ApplicationDefinition) *MetaData {
	md_lock.Lock()
	defer md_lock.Unlock()
	for _, md := range metadata {
		if md.appdef.ID == appdef.ID {
			return md
		}
	}
	md := &MetaData{appdef: appdef}
	metadata = append(metadata, md)
	return md
}
func GetMetaData(appdef *pb.ApplicationDefinition) *MetaData {
	md_lock.Lock()
	defer md_lock.Unlock()
	for _, md := range metadata {
		if md.appdef.ID == appdef.ID {
			return md
		}
	}
	return nil
}
func (md *MetaData) UserDisabled() bool {
	if md == nil {
		return false
	}
	return md.user_requested_stop
}
