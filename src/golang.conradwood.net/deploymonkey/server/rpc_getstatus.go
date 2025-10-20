package main

import (
	"context"

	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/deployq"

	common "golang.conradwood.net/apis/common"
)

func (dm *DeployMonkey) GetStatus(ctx context.Context, req *common.Void) (*pb.Status, error) {
	b := applying_suggestions || deployq.IsDeploying()
	res := &pb.Status{
		CurrentlyApplyingSuggestions: b,
	}
	return res, nil

}
