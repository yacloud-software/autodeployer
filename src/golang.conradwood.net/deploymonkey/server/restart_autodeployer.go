package main

import (
	"context"
	"golang.conradwood.net/apis/common"
)

func (s *DeployMonkey) AutodeployerShutdown(ctx context.Context, req *common.Void) (*common.Void, error) {
	return &common.Void{}, nil
}
func (s *DeployMonkey) AutodeployerStartup(ctx context.Context, req *common.Void) (*common.Void, error) {
	return &common.Void{}, nil
}
