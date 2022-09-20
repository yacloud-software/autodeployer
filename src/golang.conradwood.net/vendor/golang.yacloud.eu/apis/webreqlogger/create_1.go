// client create: WebReqLoggerClient
/* geninfo:
   filename  : golang.yacloud.eu/apis/webreqlogger/webreqlogger.proto
   gopackage : golang.yacloud.eu/apis/webreqlogger
   importname: ai_0
   varname   : client_WebReqLoggerClient_0
   clientname: WebReqLoggerClient
   servername: WebReqLoggerServer
   gscvname  : webreqlogger.WebReqLogger
   lockname  : lock_WebReqLoggerClient_0
   activename: active_WebReqLoggerClient_0
*/

package webreqlogger

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_WebReqLoggerClient_0 sync.Mutex
  client_WebReqLoggerClient_0 WebReqLoggerClient
)

func GetWebReqLoggerClient() WebReqLoggerClient { 
    if client_WebReqLoggerClient_0 != nil {
        return client_WebReqLoggerClient_0
    }

    lock_WebReqLoggerClient_0.Lock() 
    if client_WebReqLoggerClient_0 != nil {
       lock_WebReqLoggerClient_0.Unlock()
       return client_WebReqLoggerClient_0
    }

    client_WebReqLoggerClient_0 = NewWebReqLoggerClient(client.Connect("webreqlogger.WebReqLogger"))
    lock_WebReqLoggerClient_0.Unlock()
    return client_WebReqLoggerClient_0
}

