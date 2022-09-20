// client create: ShopClient
/* geninfo:
   filename  : golang.yacloud.eu/apis/shop/shop.proto
   gopackage : golang.yacloud.eu/apis/shop
   importname: ai_0
   varname   : client_ShopClient_0
   clientname: ShopClient
   servername: ShopServer
   gscvname  : shop.Shop
   lockname  : lock_ShopClient_0
   activename: active_ShopClient_0
*/

package shop

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_ShopClient_0 sync.Mutex
  client_ShopClient_0 ShopClient
)

func GetShopClient() ShopClient { 
    if client_ShopClient_0 != nil {
        return client_ShopClient_0
    }

    lock_ShopClient_0.Lock() 
    if client_ShopClient_0 != nil {
       lock_ShopClient_0.Unlock()
       return client_ShopClient_0
    }

    client_ShopClient_0 = NewShopClient(client.Connect("shop.Shop"))
    lock_ShopClient_0.Unlock()
    return client_ShopClient_0
}

