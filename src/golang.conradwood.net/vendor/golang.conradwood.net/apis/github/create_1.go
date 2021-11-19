// client create: GitHubClient
/* geninfo:
   filename  : golang.conradwood.net/apis/github/github.proto
   gopackage : golang.conradwood.net/apis/github
   importname: ai_0
   varname   : client_GitHubClient_0
   clientname: GitHubClient
   servername: GitHubServer
   gscvname  : github.GitHub
   lockname  : lock_GitHubClient_0
   activename: active_GitHubClient_0
*/

package github

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_GitHubClient_0 sync.Mutex
  client_GitHubClient_0 GitHubClient
)

func GetGitHubClient() GitHubClient { 
    if client_GitHubClient_0 != nil {
        return client_GitHubClient_0
    }

    lock_GitHubClient_0.Lock() 
    if client_GitHubClient_0 != nil {
       lock_GitHubClient_0.Unlock()
       return client_GitHubClient_0
    }

    client_GitHubClient_0 = NewGitHubClient(client.Connect("github.GitHub"))
    lock_GitHubClient_0.Unlock()
    return client_GitHubClient_0
}

