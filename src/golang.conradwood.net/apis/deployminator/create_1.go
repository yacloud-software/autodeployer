// client create: DeployminatorClient
/*
  Created by /srv/home/cnw/devel/go/go-tools/src/golang.conradwood.net/gotools/protoc-gen-cnw/protoc-gen-cnw.go
*/

/* geninfo:
   filename  : protos/golang.conradwood.net/apis/deployminator/deployminator.proto
   gopackage : golang.conradwood.net/apis/deployminator
   importname: ai_0
   clientfunc: GetDeployminator
   serverfunc: NewDeployminator
   lookupfunc: DeployminatorLookupID
   varname   : client_DeployminatorClient_0
   clientname: DeployminatorClient
   servername: DeployminatorServer
   gscvname  : deployminator.Deployminator
   lockname  : lock_DeployminatorClient_0
   activename: active_DeployminatorClient_0
*/

package deployminator

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_DeployminatorClient_0 sync.Mutex
  client_DeployminatorClient_0 DeployminatorClient
)

func GetDeployminatorClient() DeployminatorClient { 
    if client_DeployminatorClient_0 != nil {
        return client_DeployminatorClient_0
    }

    lock_DeployminatorClient_0.Lock() 
    if client_DeployminatorClient_0 != nil {
       lock_DeployminatorClient_0.Unlock()
       return client_DeployminatorClient_0
    }

    client_DeployminatorClient_0 = NewDeployminatorClient(client.Connect(DeployminatorLookupID()))
    lock_DeployminatorClient_0.Unlock()
    return client_DeployminatorClient_0
}

func DeployminatorLookupID() string { return "deployminator.Deployminator" } // returns the ID suitable for lookup in the registry. treat as opaque, subject to change.
