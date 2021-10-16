// client create: VpnManagerClient
/* geninfo:
   filename  : golang.conradwood.net/apis/vpnmanager/vpnmanager.proto
   gopackage : golang.conradwood.net/apis/vpnmanager
   importname: ai_0
   varname   : client_VpnManagerClient_0
   clientname: VpnManagerClient
   servername: VpnManagerServer
   gscvname  : vpnmanager.VpnManager
   lockname  : lock_VpnManagerClient_0
   activename: active_VpnManagerClient_0
*/

package vpnmanager

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_VpnManagerClient_0 sync.Mutex
  client_VpnManagerClient_0 VpnManagerClient
)

func GetVpnManagerClient() VpnManagerClient { 
    if client_VpnManagerClient_0 != nil {
        return client_VpnManagerClient_0
    }

    lock_VpnManagerClient_0.Lock() 
    if client_VpnManagerClient_0 != nil {
       lock_VpnManagerClient_0.Unlock()
       return client_VpnManagerClient_0
    }

    client_VpnManagerClient_0 = NewVpnManagerClient(client.Connect("vpnmanager.VpnManager"))
    lock_VpnManagerClient_0.Unlock()
    return client_VpnManagerClient_0
}

