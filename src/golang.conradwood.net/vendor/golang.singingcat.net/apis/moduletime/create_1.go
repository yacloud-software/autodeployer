// client create: ModuleTimeClient
/* geninfo:
   filename  : golang.singingcat.net/apis/moduletime/moduletime.proto
   gopackage : golang.singingcat.net/apis/moduletime
   importname: ai_0
   varname   : client_ModuleTimeClient_0
   clientname: ModuleTimeClient
   servername: ModuleTimeServer
   gscvname  : moduletime.ModuleTime
   lockname  : lock_ModuleTimeClient_0
   activename: active_ModuleTimeClient_0
*/

package moduletime

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_ModuleTimeClient_0 sync.Mutex
  client_ModuleTimeClient_0 ModuleTimeClient
)

func GetModuleTimeClient() ModuleTimeClient { 
    if client_ModuleTimeClient_0 != nil {
        return client_ModuleTimeClient_0
    }

    lock_ModuleTimeClient_0.Lock() 
    if client_ModuleTimeClient_0 != nil {
       lock_ModuleTimeClient_0.Unlock()
       return client_ModuleTimeClient_0
    }

    client_ModuleTimeClient_0 = NewModuleTimeClient(client.Connect("moduletime.ModuleTime"))
    lock_ModuleTimeClient_0.Unlock()
    return client_ModuleTimeClient_0
}

