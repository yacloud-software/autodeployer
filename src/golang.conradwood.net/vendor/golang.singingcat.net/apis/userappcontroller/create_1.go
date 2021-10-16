// client create: UserAppControllerClient
/* geninfo:
   filename  : golang.singingcat.net/apis/userappcontroller/userappcontroller.proto
   gopackage : golang.singingcat.net/apis/userappcontroller
   importname: ai_0
   varname   : client_UserAppControllerClient_0
   clientname: UserAppControllerClient
   servername: UserAppControllerServer
   gscvname  : userappcontroller.UserAppController
   lockname  : lock_UserAppControllerClient_0
   activename: active_UserAppControllerClient_0
*/

package userappcontroller

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_UserAppControllerClient_0 sync.Mutex
  client_UserAppControllerClient_0 UserAppControllerClient
)

func GetUserAppControllerClient() UserAppControllerClient { 
    if client_UserAppControllerClient_0 != nil {
        return client_UserAppControllerClient_0
    }

    lock_UserAppControllerClient_0.Lock() 
    if client_UserAppControllerClient_0 != nil {
       lock_UserAppControllerClient_0.Unlock()
       return client_UserAppControllerClient_0
    }

    client_UserAppControllerClient_0 = NewUserAppControllerClient(client.Connect("userappcontroller.UserAppController"))
    lock_UserAppControllerClient_0.Unlock()
    return client_UserAppControllerClient_0
}

