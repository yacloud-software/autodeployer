// client create: PaymentsClient
/* geninfo:
   filename  : golang.yacloud.eu/apis/payments/payments.proto
   gopackage : golang.yacloud.eu/apis/payments
   importname: ai_0
   varname   : client_PaymentsClient_0
   clientname: PaymentsClient
   servername: PaymentsServer
   gscvname  : payments.Payments
   lockname  : lock_PaymentsClient_0
   activename: active_PaymentsClient_0
*/

package payments

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_PaymentsClient_0 sync.Mutex
  client_PaymentsClient_0 PaymentsClient
)

func GetPaymentsClient() PaymentsClient { 
    if client_PaymentsClient_0 != nil {
        return client_PaymentsClient_0
    }

    lock_PaymentsClient_0.Lock() 
    if client_PaymentsClient_0 != nil {
       lock_PaymentsClient_0.Unlock()
       return client_PaymentsClient_0
    }

    client_PaymentsClient_0 = NewPaymentsClient(client.Connect("payments.Payments"))
    lock_PaymentsClient_0.Unlock()
    return client_PaymentsClient_0
}

