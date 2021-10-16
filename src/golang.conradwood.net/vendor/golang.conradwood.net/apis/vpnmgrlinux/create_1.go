// client create: VpnMgrLinuxClient
/* geninfo:
   filename  : golang.conradwood.net/apis/vpnmgrlinux/vpnmgrlinux.proto
   gopackage : golang.conradwood.net/apis/vpnmgrlinux
   importname: ai_0
   varname   : client_VpnMgrLinuxClient_0
   clientname: VpnMgrLinuxClient
   servername: VpnMgrLinuxServer
   gscvname  : vpnmgrlinux.VpnMgrLinux
   lockname  : lock_VpnMgrLinuxClient_0
   activename: active_VpnMgrLinuxClient_0
*/

package vpnmgrlinux

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_VpnMgrLinuxClient_0 sync.Mutex
  client_VpnMgrLinuxClient_0 VpnMgrLinuxClient
)

func GetVpnMgrLinuxClient() VpnMgrLinuxClient { 
    if client_VpnMgrLinuxClient_0 != nil {
        return client_VpnMgrLinuxClient_0
    }

    lock_VpnMgrLinuxClient_0.Lock() 
    if client_VpnMgrLinuxClient_0 != nil {
       lock_VpnMgrLinuxClient_0.Unlock()
       return client_VpnMgrLinuxClient_0
    }

    client_VpnMgrLinuxClient_0 = NewVpnMgrLinuxClient(client.Connect("vpnmgrlinux.VpnMgrLinux"))
    lock_VpnMgrLinuxClient_0.Unlock()
    return client_VpnMgrLinuxClient_0
}

