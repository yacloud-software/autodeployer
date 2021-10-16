// client create: RFSimulatorClient
/* geninfo:
   filename  : golang.conradwood.net/apis/rfsimulator/rfsimulator.proto
   gopackage : golang.conradwood.net/apis/rfsimulator
   importname: ai_0
   varname   : client_RFSimulatorClient_0
   clientname: RFSimulatorClient
   servername: RFSimulatorServer
   gscvname  : rfsimulator.RFSimulator
   lockname  : lock_RFSimulatorClient_0
   activename: active_RFSimulatorClient_0
*/

package rfsimulator

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_RFSimulatorClient_0 sync.Mutex
  client_RFSimulatorClient_0 RFSimulatorClient
)

func GetRFSimulatorClient() RFSimulatorClient { 
    if client_RFSimulatorClient_0 != nil {
        return client_RFSimulatorClient_0
    }

    lock_RFSimulatorClient_0.Lock() 
    if client_RFSimulatorClient_0 != nil {
       lock_RFSimulatorClient_0.Unlock()
       return client_RFSimulatorClient_0
    }

    client_RFSimulatorClient_0 = NewRFSimulatorClient(client.Connect("rfsimulator.RFSimulator"))
    lock_RFSimulatorClient_0.Unlock()
    return client_RFSimulatorClient_0
}

