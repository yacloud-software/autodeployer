// client create: ImageRecorderClient
/* geninfo:
   filename  : golang.conradwood.net/apis/imagerecorder/imagerecorder.proto
   gopackage : golang.conradwood.net/apis/imagerecorder
   importname: ai_0
   varname   : client_ImageRecorderClient_0
   clientname: ImageRecorderClient
   servername: ImageRecorderServer
   gscvname  : imagerecorder.ImageRecorder
   lockname  : lock_ImageRecorderClient_0
   activename: active_ImageRecorderClient_0
*/

package imagerecorder

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_ImageRecorderClient_0 sync.Mutex
  client_ImageRecorderClient_0 ImageRecorderClient
)

func GetImageRecorderClient() ImageRecorderClient { 
    if client_ImageRecorderClient_0 != nil {
        return client_ImageRecorderClient_0
    }

    lock_ImageRecorderClient_0.Lock() 
    if client_ImageRecorderClient_0 != nil {
       lock_ImageRecorderClient_0.Unlock()
       return client_ImageRecorderClient_0
    }

    client_ImageRecorderClient_0 = NewImageRecorderClient(client.Connect("imagerecorder.ImageRecorder"))
    lock_ImageRecorderClient_0.Unlock()
    return client_ImageRecorderClient_0
}

