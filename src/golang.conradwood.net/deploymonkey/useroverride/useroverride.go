package useroverride

import pb "golang.conradwood.net/apis/deploymonkey"

// user requested this to be undeployed (do not restart it automatically)
func MarkAsUndeployed(appdef *pb.ApplicationDefinition) {
	md := GetOrCreateMetaData(appdef)
	md.Modify(func(m *MetaData) {
		m.appmeta.UserRequestedStop = true
	})
}

// user requested this to be undeployed (do not restart it automatically)
func MarkAsDeployed(appdef *pb.ApplicationDefinition) {
	md := GetMetaData(appdef)
	if md == nil {
		return
	}
	md.Modify(func(m *MetaData) {
		m.appmeta.UserRequestedStop = false
	})
}
