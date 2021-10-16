// client create: GitDiffSyncClient
/* geninfo:
   filename  : golang.conradwood.net/apis/gitdiffsync/gitdiffsync.proto
   gopackage : golang.conradwood.net/apis/gitdiffsync
   importname: ai_0
   varname   : client_GitDiffSyncClient_0
   clientname: GitDiffSyncClient
   servername: GitDiffSyncServer
   gscvname  : gitdiffsync.GitDiffSync
   lockname  : lock_GitDiffSyncClient_0
   activename: active_GitDiffSyncClient_0
*/

package gitdiffsync

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_GitDiffSyncClient_0 sync.Mutex
  client_GitDiffSyncClient_0 GitDiffSyncClient
)

func GetGitDiffSyncClient() GitDiffSyncClient { 
    if client_GitDiffSyncClient_0 != nil {
        return client_GitDiffSyncClient_0
    }

    lock_GitDiffSyncClient_0.Lock() 
    if client_GitDiffSyncClient_0 != nil {
       lock_GitDiffSyncClient_0.Unlock()
       return client_GitDiffSyncClient_0
    }

    client_GitDiffSyncClient_0 = NewGitDiffSyncClient(client.Connect("gitdiffsync.GitDiffSync"))
    lock_GitDiffSyncClient_0.Unlock()
    return client_GitDiffSyncClient_0
}

