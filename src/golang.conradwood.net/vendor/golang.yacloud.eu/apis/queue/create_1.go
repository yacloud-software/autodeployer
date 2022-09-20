// client create: QueueClient
/* geninfo:
   filename  : golang.yacloud.eu/apis/queue/queue.proto
   gopackage : golang.yacloud.eu/apis/queue
   importname: ai_0
   varname   : client_QueueClient_0
   clientname: QueueClient
   servername: QueueServer
   gscvname  : queue.Queue
   lockname  : lock_QueueClient_0
   activename: active_QueueClient_0
*/

package queue

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_QueueClient_0 sync.Mutex
  client_QueueClient_0 QueueClient
)

func GetQueueClient() QueueClient { 
    if client_QueueClient_0 != nil {
        return client_QueueClient_0
    }

    lock_QueueClient_0.Lock() 
    if client_QueueClient_0 != nil {
       lock_QueueClient_0.Unlock()
       return client_QueueClient_0
    }

    client_QueueClient_0 = NewQueueClient(client.Connect("queue.Queue"))
    lock_QueueClient_0.Unlock()
    return client_QueueClient_0
}

