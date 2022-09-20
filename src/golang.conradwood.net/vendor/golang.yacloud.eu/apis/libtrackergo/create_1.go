// client create: LibTrackerGoClient
/* geninfo:
   filename  : golang.yacloud.eu/apis/libtrackergo/libtrackergo.proto
   gopackage : golang.yacloud.eu/apis/libtrackergo
   importname: ai_0
   varname   : client_LibTrackerGoClient_0
   clientname: LibTrackerGoClient
   servername: LibTrackerGoServer
   gscvname  : libtrackergo.LibTrackerGo
   lockname  : lock_LibTrackerGoClient_0
   activename: active_LibTrackerGoClient_0
*/

package libtrackergo

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_LibTrackerGoClient_0 sync.Mutex
  client_LibTrackerGoClient_0 LibTrackerGoClient
)

func GetLibTrackerGoClient() LibTrackerGoClient { 
    if client_LibTrackerGoClient_0 != nil {
        return client_LibTrackerGoClient_0
    }

    lock_LibTrackerGoClient_0.Lock() 
    if client_LibTrackerGoClient_0 != nil {
       lock_LibTrackerGoClient_0.Unlock()
       return client_LibTrackerGoClient_0
    }

    client_LibTrackerGoClient_0 = NewLibTrackerGoClient(client.Connect("libtrackergo.LibTrackerGo"))
    lock_LibTrackerGoClient_0.Unlock()
    return client_LibTrackerGoClient_0
}

