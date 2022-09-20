// client create: BitfolkClient
/* geninfo:
   filename  : golang.conradwood.net/apis/bitfolk/bitfolk.proto
   gopackage : golang.conradwood.net/apis/bitfolk
   importname: ai_0
   varname   : client_BitfolkClient_0
   clientname: BitfolkClient
   servername: BitfolkServer
   gscvname  : bitfolk.Bitfolk
   lockname  : lock_BitfolkClient_0
   activename: active_BitfolkClient_0
*/

package bitfolk

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_BitfolkClient_0 sync.Mutex
  client_BitfolkClient_0 BitfolkClient
)

func GetBitfolkClient() BitfolkClient { 
    if client_BitfolkClient_0 != nil {
        return client_BitfolkClient_0
    }

    lock_BitfolkClient_0.Lock() 
    if client_BitfolkClient_0 != nil {
       lock_BitfolkClient_0.Unlock()
       return client_BitfolkClient_0
    }

    client_BitfolkClient_0 = NewBitfolkClient(client.Connect("bitfolk.Bitfolk"))
    lock_BitfolkClient_0.Unlock()
    return client_BitfolkClient_0
}

