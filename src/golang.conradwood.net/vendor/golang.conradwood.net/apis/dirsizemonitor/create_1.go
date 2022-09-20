// client create: DirSizeMonitorClient
/* geninfo:
   filename  : golang.conradwood.net/apis/dirsizemonitor/dirsizemonitor.proto
   gopackage : golang.conradwood.net/apis/dirsizemonitor
   importname: ai_0
   varname   : client_DirSizeMonitorClient_0
   clientname: DirSizeMonitorClient
   servername: DirSizeMonitorServer
   gscvname  : dirsizemonitor.DirSizeMonitor
   lockname  : lock_DirSizeMonitorClient_0
   activename: active_DirSizeMonitorClient_0
*/

package dirsizemonitor

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_DirSizeMonitorClient_0 sync.Mutex
  client_DirSizeMonitorClient_0 DirSizeMonitorClient
)

func GetDirSizeMonitorClient() DirSizeMonitorClient { 
    if client_DirSizeMonitorClient_0 != nil {
        return client_DirSizeMonitorClient_0
    }

    lock_DirSizeMonitorClient_0.Lock() 
    if client_DirSizeMonitorClient_0 != nil {
       lock_DirSizeMonitorClient_0.Unlock()
       return client_DirSizeMonitorClient_0
    }

    client_DirSizeMonitorClient_0 = NewDirSizeMonitorClient(client.Connect("dirsizemonitor.DirSizeMonitor"))
    lock_DirSizeMonitorClient_0.Unlock()
    return client_DirSizeMonitorClient_0
}

