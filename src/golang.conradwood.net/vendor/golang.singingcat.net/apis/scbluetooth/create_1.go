// client create: SCBluetoothClient
/* geninfo:
   filename  : golang.singingcat.net/apis/scbluetooth/scbluetooth.proto
   gopackage : golang.singingcat.net/apis/scbluetooth
   importname: ai_0
   varname   : client_SCBluetoothClient_0
   clientname: SCBluetoothClient
   servername: SCBluetoothServer
   gscvname  : scbluetooth.SCBluetooth
   lockname  : lock_SCBluetoothClient_0
   activename: active_SCBluetoothClient_0
*/

package scbluetooth

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_SCBluetoothClient_0 sync.Mutex
  client_SCBluetoothClient_0 SCBluetoothClient
)

func GetSCBluetoothClient() SCBluetoothClient { 
    if client_SCBluetoothClient_0 != nil {
        return client_SCBluetoothClient_0
    }

    lock_SCBluetoothClient_0.Lock() 
    if client_SCBluetoothClient_0 != nil {
       lock_SCBluetoothClient_0.Unlock()
       return client_SCBluetoothClient_0
    }

    client_SCBluetoothClient_0 = NewSCBluetoothClient(client.Connect("scbluetooth.SCBluetooth"))
    lock_SCBluetoothClient_0.Unlock()
    return client_SCBluetoothClient_0
}

