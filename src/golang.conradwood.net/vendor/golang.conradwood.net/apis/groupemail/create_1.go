// client create: GroupEmailClient
/* geninfo:
   filename  : golang.conradwood.net/apis/groupemail/groupemail.proto
   gopackage : golang.conradwood.net/apis/groupemail
   importname: ai_0
   varname   : client_GroupEmailClient_0
   clientname: GroupEmailClient
   servername: GroupEmailServer
   gscvname  : groupemail.GroupEmail
   lockname  : lock_GroupEmailClient_0
   activename: active_GroupEmailClient_0
*/

package groupemail

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_GroupEmailClient_0 sync.Mutex
  client_GroupEmailClient_0 GroupEmailClient
)

func GetGroupEmailClient() GroupEmailClient { 
    if client_GroupEmailClient_0 != nil {
        return client_GroupEmailClient_0
    }

    lock_GroupEmailClient_0.Lock() 
    if client_GroupEmailClient_0 != nil {
       lock_GroupEmailClient_0.Unlock()
       return client_GroupEmailClient_0
    }

    client_GroupEmailClient_0 = NewGroupEmailClient(client.Connect("groupemail.GroupEmail"))
    lock_GroupEmailClient_0.Unlock()
    return client_GroupEmailClient_0
}

