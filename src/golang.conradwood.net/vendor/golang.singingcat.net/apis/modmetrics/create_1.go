// client create: ModMetricsClient
/* geninfo:
   filename  : golang.singingcat.net/apis/modmetrics/modmetrics.proto
   gopackage : golang.singingcat.net/apis/modmetrics
   importname: ai_0
   varname   : client_ModMetricsClient_0
   clientname: ModMetricsClient
   servername: ModMetricsServer
   gscvname  : modmetrics.ModMetrics
   lockname  : lock_ModMetricsClient_0
   activename: active_ModMetricsClient_0
*/

package modmetrics

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_ModMetricsClient_0 sync.Mutex
  client_ModMetricsClient_0 ModMetricsClient
)

func GetModMetricsClient() ModMetricsClient { 
    if client_ModMetricsClient_0 != nil {
        return client_ModMetricsClient_0
    }

    lock_ModMetricsClient_0.Lock() 
    if client_ModMetricsClient_0 != nil {
       lock_ModMetricsClient_0.Unlock()
       return client_ModMetricsClient_0
    }

    client_ModMetricsClient_0 = NewModMetricsClient(client.Connect("modmetrics.ModMetrics"))
    lock_ModMetricsClient_0.Unlock()
    return client_ModMetricsClient_0
}

