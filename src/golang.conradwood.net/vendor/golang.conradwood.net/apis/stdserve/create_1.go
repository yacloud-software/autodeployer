// client create: StdServeClient
/* geninfo:
   filename  : golang.conradwood.net/apis/stdserve/stdserve.proto
   gopackage : golang.conradwood.net/apis/stdserve
   importname: ai_0
   varname   : client_StdServeClient_0
   clientname: StdServeClient
   servername: StdServeServer
   gscvname  : stdserve.StdServe
   lockname  : lock_StdServeClient_0
   activename: active_StdServeClient_0
*/

package stdserve

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_StdServeClient_0 sync.Mutex
  client_StdServeClient_0 StdServeClient
)

func GetStdServeClient() StdServeClient { 
    if client_StdServeClient_0 != nil {
        return client_StdServeClient_0
    }

    lock_StdServeClient_0.Lock() 
    if client_StdServeClient_0 != nil {
       lock_StdServeClient_0.Unlock()
       return client_StdServeClient_0
    }

    client_StdServeClient_0 = NewStdServeClient(client.Connect("stdserve.StdServe"))
    lock_StdServeClient_0.Unlock()
    return client_StdServeClient_0
}

