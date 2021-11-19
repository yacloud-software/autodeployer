// client create: GoProxyClient
/* geninfo:
   filename  : golang.conradwood.net/apis/goproxy/goproxy.proto
   gopackage : golang.conradwood.net/apis/goproxy
   importname: ai_0
   varname   : client_GoProxyClient_0
   clientname: GoProxyClient
   servername: GoProxyServer
   gscvname  : goproxy.GoProxy
   lockname  : lock_GoProxyClient_0
   activename: active_GoProxyClient_0
*/

package goproxy

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_GoProxyClient_0 sync.Mutex
  client_GoProxyClient_0 GoProxyClient
)

func GetGoProxyClient() GoProxyClient { 
    if client_GoProxyClient_0 != nil {
        return client_GoProxyClient_0
    }

    lock_GoProxyClient_0.Lock() 
    if client_GoProxyClient_0 != nil {
       lock_GoProxyClient_0.Unlock()
       return client_GoProxyClient_0
    }

    client_GoProxyClient_0 = NewGoProxyClient(client.Connect("goproxy.GoProxy"))
    lock_GoProxyClient_0.Unlock()
    return client_GoProxyClient_0
}

