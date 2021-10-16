// client create: LocationServiceClient
/* geninfo:
   filename  : golang.conradwood.net/apis/location/location.proto
   gopackage : golang.conradwood.net/apis/location
   importname: ai_0
   varname   : client_LocationServiceClient_0
   clientname: LocationServiceClient
   servername: LocationServiceServer
   gscvname  : location.LocationService
   lockname  : lock_LocationServiceClient_0
   activename: active_LocationServiceClient_0
*/

package location

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_LocationServiceClient_0 sync.Mutex
  client_LocationServiceClient_0 LocationServiceClient
)

func GetLocationClient() LocationServiceClient { 
    if client_LocationServiceClient_0 != nil {
        return client_LocationServiceClient_0
    }

    lock_LocationServiceClient_0.Lock() 
    if client_LocationServiceClient_0 != nil {
       lock_LocationServiceClient_0.Unlock()
       return client_LocationServiceClient_0
    }

    client_LocationServiceClient_0 = NewLocationServiceClient(client.Connect("location.LocationService"))
    lock_LocationServiceClient_0.Unlock()
    return client_LocationServiceClient_0
}

func GetLocationServiceClient() LocationServiceClient { 
    if client_LocationServiceClient_0 != nil {
        return client_LocationServiceClient_0
    }

    lock_LocationServiceClient_0.Lock() 
    if client_LocationServiceClient_0 != nil {
       lock_LocationServiceClient_0.Unlock()
       return client_LocationServiceClient_0
    }

    client_LocationServiceClient_0 = NewLocationServiceClient(client.Connect("location.LocationService"))
    lock_LocationServiceClient_0.Unlock()
    return client_LocationServiceClient_0
}

