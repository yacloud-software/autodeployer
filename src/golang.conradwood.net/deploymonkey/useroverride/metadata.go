package useroverride

import (
	"context"
	"fmt"
	"sync"

	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/db"
	"golang.conradwood.net/go-easyops/errors"
)

var (
	md_lock  sync.Mutex
	metadata []*MetaData
)

type MetaData struct {
	sync.Mutex
	appdef  *pb.ApplicationDefinition
	appmeta *pb.AppMeta
}

func GetOrCreateMetaData(appdef *pb.ApplicationDefinition) *MetaData {
	return getMaybeCreateMetaData(appdef, true)
}
func GetMetaData(appdef *pb.ApplicationDefinition) *MetaData {
	return getMaybeCreateMetaData(appdef, false)
}

func getMaybeCreateMetaData(appdef *pb.ApplicationDefinition, autocreate bool) *MetaData {
	md_lock.Lock()
	defer md_lock.Unlock()
	for _, md := range metadata {
		if md.appdef.ID == appdef.ID {
			return md
		}
	}
	ams, err := db.DefaultDBAppMeta().All(context.Background())
	if err != nil {
		fmt.Printf("failed to get metas: %s\n", errors.ErrorString(err))
		return nil
	}
	for _, am := range ams {
		if am.AppDef.ID == appdef.ID {
			md := &MetaData{appdef: appdef, appmeta: am}
			metadata = append(metadata, md)
			return md
		}
	}
	if !autocreate {
		return nil
	}
	md := &MetaData{appdef: appdef}
	metadata = append(metadata, md)
	return md
}
func (md *MetaData) UserDisabled() bool {
	if md == nil || md.appmeta == nil {
		return false
	}
	return md.appmeta.UserRequestedStop
}
func (md *MetaData) Modify(f func(md *MetaData)) error {
	md.Lock()
	if md.appmeta == nil {
		md.appmeta = &pb.AppMeta{AppDef: md.appdef}
	}
	f(md)
	md.Unlock()
	return md.save()
}
func (md *MetaData) save() error {
	md.Lock()
	defer md.Unlock()
	if md.appmeta == nil {
		return nil
	}
	var err error
	if md.appmeta.ID == 0 {
		_, err = db.DefaultDBAppMeta().Save(context.Background(), md.appmeta)
	} else {
		err = db.DefaultDBAppMeta().Update(context.Background(), md.appmeta)
	}
	return err
}
