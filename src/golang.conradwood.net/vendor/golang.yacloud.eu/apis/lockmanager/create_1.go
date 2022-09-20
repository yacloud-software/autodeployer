// client create: LockManagerClient
/* geninfo:
   filename  : golang.yacloud.eu/apis/lockmanager/lockmanager.proto
   gopackage : golang.yacloud.eu/apis/lockmanager
   importname: ai_0
   varname   : client_LockManagerClient_0
   clientname: LockManagerClient
   servername: LockManagerServer
   gscvname  : lockmanager.LockManager
   lockname  : lock_LockManagerClient_0
   activename: active_LockManagerClient_0
*/

package lockmanager

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_LockManagerClient_0 sync.Mutex
  client_LockManagerClient_0 LockManagerClient
)

func GetLockManagerClient() LockManagerClient { 
    if client_LockManagerClient_0 != nil {
        return client_LockManagerClient_0
    }

    lock_LockManagerClient_0.Lock() 
    if client_LockManagerClient_0 != nil {
       lock_LockManagerClient_0.Unlock()
       return client_LockManagerClient_0
    }

    client_LockManagerClient_0 = NewLockManagerClient(client.Connect("lockmanager.LockManager"))
    lock_LockManagerClient_0.Unlock()
    return client_LockManagerClient_0
}

