// client create: BuildRepoArchiveClient
/* geninfo:
   filename  : golang.yacloud.eu/apis/buildrepoarchive/buildrepoarchive.proto
   gopackage : golang.yacloud.eu/apis/buildrepoarchive
   importname: ai_0
   varname   : client_BuildRepoArchiveClient_0
   clientname: BuildRepoArchiveClient
   servername: BuildRepoArchiveServer
   gscvname  : buildrepoarchive.BuildRepoArchive
   lockname  : lock_BuildRepoArchiveClient_0
   activename: active_BuildRepoArchiveClient_0
*/

package buildrepoarchive

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_BuildRepoArchiveClient_0 sync.Mutex
  client_BuildRepoArchiveClient_0 BuildRepoArchiveClient
)

func GetBuildRepoArchiveClient() BuildRepoArchiveClient { 
    if client_BuildRepoArchiveClient_0 != nil {
        return client_BuildRepoArchiveClient_0
    }

    lock_BuildRepoArchiveClient_0.Lock() 
    if client_BuildRepoArchiveClient_0 != nil {
       lock_BuildRepoArchiveClient_0.Unlock()
       return client_BuildRepoArchiveClient_0
    }

    client_BuildRepoArchiveClient_0 = NewBuildRepoArchiveClient(client.Connect("buildrepoarchive.BuildRepoArchive"))
    lock_BuildRepoArchiveClient_0.Unlock()
    return client_BuildRepoArchiveClient_0
}

