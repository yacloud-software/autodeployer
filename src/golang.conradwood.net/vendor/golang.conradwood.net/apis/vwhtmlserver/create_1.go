// client create: VWHtmlServerClient
/* geninfo:
   filename  : golang.conradwood.net/apis/vwhtmlserver/vwhtmlserver.proto
   gopackage : golang.conradwood.net/apis/vwhtmlserver
   importname: ai_0
   varname   : client_VWHtmlServerClient_0
   clientname: VWHtmlServerClient
   servername: VWHtmlServerServer
   gscvname  : vwhtmlserver.VWHtmlServer
   lockname  : lock_VWHtmlServerClient_0
   activename: active_VWHtmlServerClient_0
*/

package vwhtmlserver

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_VWHtmlServerClient_0 sync.Mutex
  client_VWHtmlServerClient_0 VWHtmlServerClient
)

func GetVWHtmlServerClient() VWHtmlServerClient { 
    if client_VWHtmlServerClient_0 != nil {
        return client_VWHtmlServerClient_0
    }

    lock_VWHtmlServerClient_0.Lock() 
    if client_VWHtmlServerClient_0 != nil {
       lock_VWHtmlServerClient_0.Unlock()
       return client_VWHtmlServerClient_0
    }

    client_VWHtmlServerClient_0 = NewVWHtmlServerClient(client.Connect("vwhtmlserver.VWHtmlServer"))
    lock_VWHtmlServerClient_0.Unlock()
    return client_VWHtmlServerClient_0
}

