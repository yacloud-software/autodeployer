package main

import (
	"flag"
	"fmt"
	"golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/echoservice"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
	"time"
)

func main() {
	flag.Parse()
	for {

		ctx := authremote.Context()
		response, err := pb.GetEchoServiceClient().Ping(ctx, &common.Void{})
		utils.Bail("failed to ping", err)
		fmt.Printf("Response: %#v\n", response)
		time.Sleep(1 * time.Second)
	}
}
