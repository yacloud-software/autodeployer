// client create: SCUsbClient
/* geninfo:
   filename  : golang.singingcat.net/apis/scusb/scusb.proto
   gopackage : golang.singingcat.net/apis/scusb
   importname: ai_0
   varname   : client_SCUsbClient_0
   clientname: SCUsbClient
   servername: SCUsbServer
   gscvname  : scusb.SCUsb
   lockname  : lock_SCUsbClient_0
   activename: active_SCUsbClient_0
*/

package scusb

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_SCUsbClient_0 sync.Mutex
  client_SCUsbClient_0 SCUsbClient
)

func GetSCUsbClient() SCUsbClient { 
    if client_SCUsbClient_0 != nil {
        return client_SCUsbClient_0
    }

    lock_SCUsbClient_0.Lock() 
    if client_SCUsbClient_0 != nil {
       lock_SCUsbClient_0.Unlock()
       return client_SCUsbClient_0
    }

    client_SCUsbClient_0 = NewSCUsbClient(client.Connect("scusb.SCUsb"))
    lock_SCUsbClient_0.Unlock()
    return client_SCUsbClient_0
}

