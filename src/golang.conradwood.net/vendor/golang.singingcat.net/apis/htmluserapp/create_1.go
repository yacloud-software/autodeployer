// client create: HTMLUserAppClient
/* geninfo:
   filename  : golang.singingcat.net/apis/htmluserapp/htmluserapp.proto
   gopackage : golang.singingcat.net/apis/htmluserapp
   importname: ai_0
   varname   : client_HTMLUserAppClient_0
   clientname: HTMLUserAppClient
   servername: HTMLUserAppServer
   gscvname  : htmluserapp.HTMLUserApp
   lockname  : lock_HTMLUserAppClient_0
   activename: active_HTMLUserAppClient_0
*/

package htmluserapp

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_HTMLUserAppClient_0 sync.Mutex
  client_HTMLUserAppClient_0 HTMLUserAppClient
)

func GetHTMLUserAppClient() HTMLUserAppClient { 
    if client_HTMLUserAppClient_0 != nil {
        return client_HTMLUserAppClient_0
    }

    lock_HTMLUserAppClient_0.Lock() 
    if client_HTMLUserAppClient_0 != nil {
       lock_HTMLUserAppClient_0.Unlock()
       return client_HTMLUserAppClient_0
    }

    client_HTMLUserAppClient_0 = NewHTMLUserAppClient(client.Connect("htmluserapp.HTMLUserApp"))
    lock_HTMLUserAppClient_0.Unlock()
    return client_HTMLUserAppClient_0
}

