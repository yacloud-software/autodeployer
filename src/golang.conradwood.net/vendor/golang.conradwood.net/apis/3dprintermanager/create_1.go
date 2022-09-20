// client create: ThreeDPrinterClient
/* geninfo:
   filename  : golang.conradwood.net/apis/3dprintermanager/3dprintermanager.proto
   gopackage : golang.conradwood.net/apis/3dprintermanager
   importname: ai_0
   varname   : client_ThreeDPrinterClient_0
   clientname: ThreeDPrinterClient
   servername: ThreeDPrinterServer
   gscvname  : threedprintermanager.ThreeDPrinter
   lockname  : lock_ThreeDPrinterClient_0
   activename: active_ThreeDPrinterClient_0
*/

package threedprintermanager

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_ThreeDPrinterClient_0 sync.Mutex
  client_ThreeDPrinterClient_0 ThreeDPrinterClient
)

func GetThreeDPrinterClient() ThreeDPrinterClient { 
    if client_ThreeDPrinterClient_0 != nil {
        return client_ThreeDPrinterClient_0
    }

    lock_ThreeDPrinterClient_0.Lock() 
    if client_ThreeDPrinterClient_0 != nil {
       lock_ThreeDPrinterClient_0.Unlock()
       return client_ThreeDPrinterClient_0
    }

    client_ThreeDPrinterClient_0 = NewThreeDPrinterClient(client.Connect("threedprintermanager.ThreeDPrinter"))
    lock_ThreeDPrinterClient_0.Unlock()
    return client_ThreeDPrinterClient_0
}

