package useroverride

import pb "golang.conradwood.net/apis/deploymonkey"

// user requested this to be undeployed (do not restart it automatically)
func MarkAsUndeployed(appdef *pb.ApplicationDefinition) {
	md := GetOrCreateMetaData(appdef)
	md.user_requested_stop = true
}

// user requested this to be undeployed (do not restart it automatically)
func MarkAsDeployed(appdef *pb.ApplicationDefinition) {
	md := GetMetaData(appdef)
	if md == nil {
		return
	}
	md.user_requested_stop = false
}
