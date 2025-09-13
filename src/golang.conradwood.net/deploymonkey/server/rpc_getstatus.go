package main

import (
	"context"

	pb "golang.conradwood.net/apis/deploymonkey"

	common "golang.conradwood.net/apis/common"
)

func (dm *DeployMonkey) GetStatus(ctx context.Context, req *common.Void) (*pb.Status, error) {
	res := &pb.Status{
		CurrentlyApplyingSuggestions: applying_suggestions,
	}
	return res, nil

}
