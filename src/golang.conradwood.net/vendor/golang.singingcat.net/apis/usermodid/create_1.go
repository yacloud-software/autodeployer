// client create: UserModIDClient
/* geninfo:
   filename  : golang.singingcat.net/apis/usermodid/usermodid.proto
   gopackage : golang.singingcat.net/apis/usermodid
   importname: ai_0
   varname   : client_UserModIDClient_0
   clientname: UserModIDClient
   servername: UserModIDServer
   gscvname  : usermodid.UserModID
   lockname  : lock_UserModIDClient_0
   activename: active_UserModIDClient_0
*/

package usermodid

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_UserModIDClient_0 sync.Mutex
  client_UserModIDClient_0 UserModIDClient
)

func GetUserModIDClient() UserModIDClient { 
    if client_UserModIDClient_0 != nil {
        return client_UserModIDClient_0
    }

    lock_UserModIDClient_0.Lock() 
    if client_UserModIDClient_0 != nil {
       lock_UserModIDClient_0.Unlock()
       return client_UserModIDClient_0
    }

    client_UserModIDClient_0 = NewUserModIDClient(client.Connect("usermodid.UserModID"))
    lock_UserModIDClient_0.Unlock()
    return client_UserModIDClient_0
}

