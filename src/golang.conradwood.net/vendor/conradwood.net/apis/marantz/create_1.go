// client create: MarantzClient
/* geninfo:
   filename  : conradwood.net/apis/marantz/marantz.proto
   gopackage : conradwood.net/apis/marantz
   importname: ai_0
   varname   : client_MarantzClient_0
   clientname: MarantzClient
   servername: MarantzServer
   gscvname  : marantz.Marantz
   lockname  : lock_MarantzClient_0
   activename: active_MarantzClient_0
*/

package marantz

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_MarantzClient_0 sync.Mutex
  client_MarantzClient_0 MarantzClient
)

func GetMarantzClient() MarantzClient { 
    if client_MarantzClient_0 != nil {
        return client_MarantzClient_0
    }

    lock_MarantzClient_0.Lock() 
    if client_MarantzClient_0 != nil {
       lock_MarantzClient_0.Unlock()
       return client_MarantzClient_0
    }

    client_MarantzClient_0 = NewMarantzClient(client.Connect("marantz.Marantz"))
    lock_MarantzClient_0.Unlock()
    return client_MarantzClient_0
}

