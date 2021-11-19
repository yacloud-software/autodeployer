// client create: SpamTrackerClient
/* geninfo:
   filename  : golang.conradwood.net/apis/spamtracker/spamtracker.proto
   gopackage : golang.conradwood.net/apis/spamtracker
   importname: ai_0
   varname   : client_SpamTrackerClient_0
   clientname: SpamTrackerClient
   servername: SpamTrackerServer
   gscvname  : spamtracker.SpamTracker
   lockname  : lock_SpamTrackerClient_0
   activename: active_SpamTrackerClient_0
*/

package spamtracker

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_SpamTrackerClient_0 sync.Mutex
  client_SpamTrackerClient_0 SpamTrackerClient
)

func GetSpamTrackerClient() SpamTrackerClient { 
    if client_SpamTrackerClient_0 != nil {
        return client_SpamTrackerClient_0
    }

    lock_SpamTrackerClient_0.Lock() 
    if client_SpamTrackerClient_0 != nil {
       lock_SpamTrackerClient_0.Unlock()
       return client_SpamTrackerClient_0
    }

    client_SpamTrackerClient_0 = NewSpamTrackerClient(client.Connect("spamtracker.SpamTracker"))
    lock_SpamTrackerClient_0.Unlock()
    return client_SpamTrackerClient_0
}

