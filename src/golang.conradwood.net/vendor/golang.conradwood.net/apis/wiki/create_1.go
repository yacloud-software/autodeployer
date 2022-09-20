// client create: WikiClient
/* geninfo:
   filename  : golang.conradwood.net/apis/wiki/wiki.proto
   gopackage : golang.conradwood.net/apis/wiki
   importname: ai_0
   varname   : client_WikiClient_0
   clientname: WikiClient
   servername: WikiServer
   gscvname  : wiki.Wiki
   lockname  : lock_WikiClient_0
   activename: active_WikiClient_0
*/

package wiki

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_WikiClient_0 sync.Mutex
  client_WikiClient_0 WikiClient
)

func GetWikiClient() WikiClient { 
    if client_WikiClient_0 != nil {
        return client_WikiClient_0
    }

    lock_WikiClient_0.Lock() 
    if client_WikiClient_0 != nil {
       lock_WikiClient_0.Unlock()
       return client_WikiClient_0
    }

    client_WikiClient_0 = NewWikiClient(client.Connect("wiki.Wiki"))
    lock_WikiClient_0.Unlock()
    return client_WikiClient_0
}

