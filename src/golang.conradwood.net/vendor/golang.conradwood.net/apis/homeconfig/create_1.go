// client create: HomeConfigClient
/* geninfo:
   filename  : golang.conradwood.net/apis/homeconfig/homeconfig.proto
   gopackage : golang.conradwood.net/apis/homeconfig
   importname: ai_0
   varname   : client_HomeConfigClient_0
   clientname: HomeConfigClient
   servername: HomeConfigServer
   gscvname  : homeconfig.HomeConfig
   lockname  : lock_HomeConfigClient_0
   activename: active_HomeConfigClient_0
*/

package homeconfig

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_HomeConfigClient_0 sync.Mutex
  client_HomeConfigClient_0 HomeConfigClient
)

func GetHomeConfigClient() HomeConfigClient { 
    if client_HomeConfigClient_0 != nil {
        return client_HomeConfigClient_0
    }

    lock_HomeConfigClient_0.Lock() 
    if client_HomeConfigClient_0 != nil {
       lock_HomeConfigClient_0.Unlock()
       return client_HomeConfigClient_0
    }

    client_HomeConfigClient_0 = NewHomeConfigClient(client.Connect("homeconfig.HomeConfig"))
    lock_HomeConfigClient_0.Unlock()
    return client_HomeConfigClient_0
}

