// client create: AssetGroupsServiceClient
/* geninfo:
   filename  : golang.conradwood.net/apis/assetgroups/assetgroups.proto
   gopackage : golang.conradwood.net/apis/assetgroups
   importname: ai_0
   varname   : client_AssetGroupsServiceClient_0
   clientname: AssetGroupsServiceClient
   servername: AssetGroupsServiceServer
   gscvname  : assetgroups.AssetGroupsService
   lockname  : lock_AssetGroupsServiceClient_0
   activename: active_AssetGroupsServiceClient_0
*/

package assetgroups

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_AssetGroupsServiceClient_0 sync.Mutex
  client_AssetGroupsServiceClient_0 AssetGroupsServiceClient
)

func GetAssetGroupsClient() AssetGroupsServiceClient { 
    if client_AssetGroupsServiceClient_0 != nil {
        return client_AssetGroupsServiceClient_0
    }

    lock_AssetGroupsServiceClient_0.Lock() 
    if client_AssetGroupsServiceClient_0 != nil {
       lock_AssetGroupsServiceClient_0.Unlock()
       return client_AssetGroupsServiceClient_0
    }

    client_AssetGroupsServiceClient_0 = NewAssetGroupsServiceClient(client.Connect("assetgroups.AssetGroupsService"))
    lock_AssetGroupsServiceClient_0.Unlock()
    return client_AssetGroupsServiceClient_0
}

func GetAssetGroupsServiceClient() AssetGroupsServiceClient { 
    if client_AssetGroupsServiceClient_0 != nil {
        return client_AssetGroupsServiceClient_0
    }

    lock_AssetGroupsServiceClient_0.Lock() 
    if client_AssetGroupsServiceClient_0 != nil {
       lock_AssetGroupsServiceClient_0.Unlock()
       return client_AssetGroupsServiceClient_0
    }

    client_AssetGroupsServiceClient_0 = NewAssetGroupsServiceClient(client.Connect("assetgroups.AssetGroupsService"))
    lock_AssetGroupsServiceClient_0.Unlock()
    return client_AssetGroupsServiceClient_0
}

