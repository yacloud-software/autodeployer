// client create: SoundClient
/* geninfo:
   filename  : golang.conradwood.net/apis/soundservice/soundservice.proto
   gopackage : golang.conradwood.net/apis/soundservice
   importname: ai_0
   varname   : client_SoundClient_0
   clientname: SoundClient
   servername: SoundServer
   gscvname  : soundservice.Sound
   lockname  : lock_SoundClient_0
   activename: active_SoundClient_0
*/

package soundservice

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_SoundClient_0 sync.Mutex
  client_SoundClient_0 SoundClient
)

func GetSoundClient() SoundClient { 
    if client_SoundClient_0 != nil {
        return client_SoundClient_0
    }

    lock_SoundClient_0.Lock() 
    if client_SoundClient_0 != nil {
       lock_SoundClient_0.Unlock()
       return client_SoundClient_0
    }

    client_SoundClient_0 = NewSoundClient(client.Connect("soundservice.Sound"))
    lock_SoundClient_0.Unlock()
    return client_SoundClient_0
}

