// client create: GoogleCastClient
/* geninfo:
   filename  : golang.conradwood.net/apis/googlecast/googlecast.proto
   gopackage : golang.conradwood.net/apis/googlecast
   importname: ai_0
   varname   : client_GoogleCastClient_0
   clientname: GoogleCastClient
   servername: GoogleCastServer
   gscvname  : googlecast.GoogleCast
   lockname  : lock_GoogleCastClient_0
   activename: active_GoogleCastClient_0
*/

package googlecast

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_GoogleCastClient_0 sync.Mutex
  client_GoogleCastClient_0 GoogleCastClient
)

func GetGoogleCastClient() GoogleCastClient { 
    if client_GoogleCastClient_0 != nil {
        return client_GoogleCastClient_0
    }

    lock_GoogleCastClient_0.Lock() 
    if client_GoogleCastClient_0 != nil {
       lock_GoogleCastClient_0.Unlock()
       return client_GoogleCastClient_0
    }

    client_GoogleCastClient_0 = NewGoogleCastClient(client.Connect("googlecast.GoogleCast"))
    lock_GoogleCastClient_0.Unlock()
    return client_GoogleCastClient_0
}

