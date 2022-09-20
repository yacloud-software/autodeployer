// client create: SCUserAppServiceClient
/* geninfo:
   filename  : users.singingcat.net/USERID/REPOSITORYNAME/REPOSITORYNAME.proto
   gopackage : users.singingcat.net/USERID/REPOSITORYNAME
   importname: ai_0
   varname   : client_SCUserAppServiceClient_0
   clientname: SCUserAppServiceClient
   servername: SCUserAppServiceServer
   gscvname  : scuserapp.SCUserAppService
   lockname  : lock_SCUserAppServiceClient_0
   activename: active_SCUserAppServiceClient_0
*/

package scuserapp

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_SCUserAppServiceClient_0 sync.Mutex
  client_SCUserAppServiceClient_0 SCUserAppServiceClient
)

func GetSCUserAppClient() SCUserAppServiceClient { 
    if client_SCUserAppServiceClient_0 != nil {
        return client_SCUserAppServiceClient_0
    }

    lock_SCUserAppServiceClient_0.Lock() 
    if client_SCUserAppServiceClient_0 != nil {
       lock_SCUserAppServiceClient_0.Unlock()
       return client_SCUserAppServiceClient_0
    }

    client_SCUserAppServiceClient_0 = NewSCUserAppServiceClient(client.Connect("scuserapp.SCUserAppService"))
    lock_SCUserAppServiceClient_0.Unlock()
    return client_SCUserAppServiceClient_0
}

func GetSCUserAppServiceClient() SCUserAppServiceClient { 
    if client_SCUserAppServiceClient_0 != nil {
        return client_SCUserAppServiceClient_0
    }

    lock_SCUserAppServiceClient_0.Lock() 
    if client_SCUserAppServiceClient_0 != nil {
       lock_SCUserAppServiceClient_0.Unlock()
       return client_SCUserAppServiceClient_0
    }

    client_SCUserAppServiceClient_0 = NewSCUserAppServiceClient(client.Connect("scuserapp.SCUserAppService"))
    lock_SCUserAppServiceClient_0.Unlock()
    return client_SCUserAppServiceClient_0
}

