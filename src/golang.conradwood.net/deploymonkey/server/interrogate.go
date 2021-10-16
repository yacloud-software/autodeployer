package main

import (
	ad "golang.conradwood.net/apis/autodeployer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// get deployments on a given host:
func getDeploymentsOnHost(ctx context.Context, conn *grpc.ClientConn) (*ad.InfoResponse, error) {
	adc := ad.NewAutoDeployerClient(conn)
	res, err := adc.GetDeployments(ctx, &ad.InfoRequest{})
	return res, err
}
