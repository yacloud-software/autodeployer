// client create: FooClient
/* geninfo:
   filename  : yacloud.eu/apis/fooname/fooname.proto
   gopackage : yacloud.eu/apis/fooname
   importname: ai_0
   varname   : client_FooClient_0
   clientname: FooClient
   servername: FooServer
   gscvname  : fooname.Foo
   lockname  : lock_FooClient_0
   activename: active_FooClient_0
*/

package fooname

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_FooClient_0 sync.Mutex
  client_FooClient_0 FooClient
)

func GetFooClient() FooClient { 
    if client_FooClient_0 != nil {
        return client_FooClient_0
    }

    lock_FooClient_0.Lock() 
    if client_FooClient_0 != nil {
       lock_FooClient_0.Unlock()
       return client_FooClient_0
    }

    client_FooClient_0 = NewFooClient(client.Connect("fooname.Foo"))
    lock_FooClient_0.Unlock()
    return client_FooClient_0
}

