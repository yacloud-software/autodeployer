// client create: SessionManagerClient
/* geninfo:
   filename  : golang.yacloud.eu/apis/sessionmanager/sessionmanager.proto
   gopackage : golang.yacloud.eu/apis/sessionmanager
   importname: ai_0
   varname   : client_SessionManagerClient_0
   clientname: SessionManagerClient
   servername: SessionManagerServer
   gscvname  : sessionmanager.SessionManager
   lockname  : lock_SessionManagerClient_0
   activename: active_SessionManagerClient_0
*/

package sessionmanager

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_SessionManagerClient_0 sync.Mutex
  client_SessionManagerClient_0 SessionManagerClient
)

func GetSessionManagerClient() SessionManagerClient { 
    if client_SessionManagerClient_0 != nil {
        return client_SessionManagerClient_0
    }

    lock_SessionManagerClient_0.Lock() 
    if client_SessionManagerClient_0 != nil {
       lock_SessionManagerClient_0.Unlock()
       return client_SessionManagerClient_0
    }

    client_SessionManagerClient_0 = NewSessionManagerClient(client.Connect("sessionmanager.SessionManager"))
    lock_SessionManagerClient_0.Unlock()
    return client_SessionManagerClient_0
}

