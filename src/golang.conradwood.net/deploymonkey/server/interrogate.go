package main

import (
	"context"
	ad "golang.conradwood.net/apis/autodeployer"
	dc "golang.conradwood.net/deploymonkey/common"
	"google.golang.org/grpc"
)

// get deployments on a given host:
func getDeploymentsOnHost(ctx context.Context, conn *grpc.ClientConn) (*ad.InfoResponse, error) {
	adc := ad.NewAutoDeployerClient(conn)
	res, err := adc.GetDeployments(ctx, dc.CreateInfoRequest())
	return res, err
}
