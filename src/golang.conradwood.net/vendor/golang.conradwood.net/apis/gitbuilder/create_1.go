// client create: GitBuilderClient
/* geninfo:
   filename  : golang.conradwood.net/apis/gitbuilder/gitbuilder.proto
   gopackage : golang.conradwood.net/apis/gitbuilder
   importname: ai_0
   varname   : client_GitBuilderClient_0
   clientname: GitBuilderClient
   servername: GitBuilderServer
   gscvname  : gitbuilder.GitBuilder
   lockname  : lock_GitBuilderClient_0
   activename: active_GitBuilderClient_0
*/

package gitbuilder

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_GitBuilderClient_0 sync.Mutex
  client_GitBuilderClient_0 GitBuilderClient
)

func GetGitBuilderClient() GitBuilderClient { 
    if client_GitBuilderClient_0 != nil {
        return client_GitBuilderClient_0
    }

    lock_GitBuilderClient_0.Lock() 
    if client_GitBuilderClient_0 != nil {
       lock_GitBuilderClient_0.Unlock()
       return client_GitBuilderClient_0
    }

    client_GitBuilderClient_0 = NewGitBuilderClient(client.Connect("gitbuilder.GitBuilder"))
    lock_GitBuilderClient_0.Unlock()
    return client_GitBuilderClient_0
}

