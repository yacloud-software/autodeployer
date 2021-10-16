// client create: AclServiceClient
/* geninfo:
   filename  : golang.conradwood.net/apis/acl/acl.proto
   gopackage : golang.conradwood.net/apis/acl
   importname: ai_0
   varname   : client_AclServiceClient_0
   clientname: AclServiceClient
   servername: AclServiceServer
   gscvname  : acl.AclService
   lockname  : lock_AclServiceClient_0
   activename: active_AclServiceClient_0
*/

package acl

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_AclServiceClient_0 sync.Mutex
  client_AclServiceClient_0 AclServiceClient
)

func GetAclClient() AclServiceClient { 
    if client_AclServiceClient_0 != nil {
        return client_AclServiceClient_0
    }

    lock_AclServiceClient_0.Lock() 
    if client_AclServiceClient_0 != nil {
       lock_AclServiceClient_0.Unlock()
       return client_AclServiceClient_0
    }

    client_AclServiceClient_0 = NewAclServiceClient(client.Connect("acl.AclService"))
    lock_AclServiceClient_0.Unlock()
    return client_AclServiceClient_0
}

func GetAclServiceClient() AclServiceClient { 
    if client_AclServiceClient_0 != nil {
        return client_AclServiceClient_0
    }

    lock_AclServiceClient_0.Lock() 
    if client_AclServiceClient_0 != nil {
       lock_AclServiceClient_0.Unlock()
       return client_AclServiceClient_0
    }

    client_AclServiceClient_0 = NewAclServiceClient(client.Connect("acl.AclService"))
    lock_AclServiceClient_0.Unlock()
    return client_AclServiceClient_0
}

