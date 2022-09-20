// client create: ObjectStoreArchiveClient
/* geninfo:
   filename  : golang.conradwood.net/apis/objectstorearchive/objectstorearchive.proto
   gopackage : golang.conradwood.net/apis/objectstorearchive
   importname: ai_0
   varname   : client_ObjectStoreArchiveClient_0
   clientname: ObjectStoreArchiveClient
   servername: ObjectStoreArchiveServer
   gscvname  : objectstorearchive.ObjectStoreArchive
   lockname  : lock_ObjectStoreArchiveClient_0
   activename: active_ObjectStoreArchiveClient_0
*/

package objectstorearchive

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_ObjectStoreArchiveClient_0 sync.Mutex
  client_ObjectStoreArchiveClient_0 ObjectStoreArchiveClient
)

func GetObjectStoreArchiveClient() ObjectStoreArchiveClient { 
    if client_ObjectStoreArchiveClient_0 != nil {
        return client_ObjectStoreArchiveClient_0
    }

    lock_ObjectStoreArchiveClient_0.Lock() 
    if client_ObjectStoreArchiveClient_0 != nil {
       lock_ObjectStoreArchiveClient_0.Unlock()
       return client_ObjectStoreArchiveClient_0
    }

    client_ObjectStoreArchiveClient_0 = NewObjectStoreArchiveClient(client.Connect("objectstorearchive.ObjectStoreArchive"))
    lock_ObjectStoreArchiveClient_0.Unlock()
    return client_ObjectStoreArchiveClient_0
}

