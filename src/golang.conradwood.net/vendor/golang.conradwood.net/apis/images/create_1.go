// client create: ImagesClient
/* geninfo:
   filename  : golang.conradwood.net/apis/images/images.proto
   gopackage : golang.conradwood.net/apis/images
   importname: ai_0
   varname   : client_ImagesClient_0
   clientname: ImagesClient
   servername: ImagesServer
   gscvname  : images.Images
   lockname  : lock_ImagesClient_0
   activename: active_ImagesClient_0
*/

package images

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_ImagesClient_0 sync.Mutex
  client_ImagesClient_0 ImagesClient
)

func GetImagesClient() ImagesClient { 
    if client_ImagesClient_0 != nil {
        return client_ImagesClient_0
    }

    lock_ImagesClient_0.Lock() 
    if client_ImagesClient_0 != nil {
       lock_ImagesClient_0.Unlock()
       return client_ImagesClient_0
    }

    client_ImagesClient_0 = NewImagesClient(client.Connect("images.Images"))
    lock_ImagesClient_0.Unlock()
    return client_ImagesClient_0
}

