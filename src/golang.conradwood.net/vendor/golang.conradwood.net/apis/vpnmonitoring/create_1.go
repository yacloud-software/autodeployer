// client create: VpnMonitoringClient
/* geninfo:
   filename  : golang.conradwood.net/apis/vpnmonitoring/vpnmonitoring.proto
   gopackage : golang.conradwood.net/apis/vpnmonitoring
   importname: ai_0
   varname   : client_VpnMonitoringClient_0
   clientname: VpnMonitoringClient
   servername: VpnMonitoringServer
   gscvname  : vpnmonitoring.VpnMonitoring
   lockname  : lock_VpnMonitoringClient_0
   activename: active_VpnMonitoringClient_0
*/

package vpnmonitoring

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_VpnMonitoringClient_0 sync.Mutex
  client_VpnMonitoringClient_0 VpnMonitoringClient
)

func GetVpnMonitoringClient() VpnMonitoringClient { 
    if client_VpnMonitoringClient_0 != nil {
        return client_VpnMonitoringClient_0
    }

    lock_VpnMonitoringClient_0.Lock() 
    if client_VpnMonitoringClient_0 != nil {
       lock_VpnMonitoringClient_0.Unlock()
       return client_VpnMonitoringClient_0
    }

    client_VpnMonitoringClient_0 = NewVpnMonitoringClient(client.Connect("vpnmonitoring.VpnMonitoring"))
    lock_VpnMonitoringClient_0.Unlock()
    return client_VpnMonitoringClient_0
}

