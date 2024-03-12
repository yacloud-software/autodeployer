package main

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/apis/grafanadata"
)

func (e *DeployMonkey) QueryTimeseries(req *grafanadata.QueryRequest, srv pb.DeployMonkey_QueryTimeseriesServer) error {
	fmt.Printf("Query timeseries: \"%s\"\n", req.Query)
	var err error
	var dps []*grafanadata.DataPoint
	ctx := srv.Context()
	if req.Query == "deployment_count" {
		dps, err = query_deployment_count(ctx, req)
	} else if req.Query == "deployment_history" {
		dps, err = query_deployment_history(ctx, req)
	}
	if err != nil {
		return err
	}
	for _, dp := range dps {
		err = srv.Send(&grafanadata.QueryResponse{DataPoint: dp})
		if err != nil {
			return err
		}
	}
	return nil
}

func query_deployment_count(ctx context.Context, req *grafanadata.QueryRequest) ([]*grafanadata.DataPoint, error) {
	return nil, fmt.Errorf("not implemented")
}
func query_deployment_history(ctx context.Context, req *grafanadata.QueryRequest) ([]*grafanadata.DataPoint, error) {
	return nil, fmt.Errorf("not implemented")
}
