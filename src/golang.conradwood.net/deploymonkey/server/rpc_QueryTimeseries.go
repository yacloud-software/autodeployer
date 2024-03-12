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
	} else if req.Query == "version_history" {
		dps, err = query_deployment_history(ctx, req)
	} else {
		return fmt.Errorf("deploymonkey does not implement query \"%s\"", req.Query)
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
	sc := GetLastQueryResult()
	ct := 0
	for _, d := range sc.deployments {
		ct = ct + len(d.Apps)
	}
	dp := &grafanadata.DataPoint{
		Timestamp: req.End,
		Value:     float64(ct),
	}
	return []*grafanadata.DataPoint{dp}, nil
}
func query_deployment_history(ctx context.Context, req *grafanadata.QueryRequest) ([]*grafanadata.DataPoint, error) {
	apps, err := appdef_store.All(ctx)
	if err != nil {
		return nil, err
	}
	var res []*grafanadata.DataPoint
	for _, app := range apps {
		if app.Created < req.Start || app.Created > req.End {
			continue
		}
		dp := &grafanadata.DataPoint{
			Timestamp: app.Created,
			Value:     float64(app.ID),
			Labels:    map[string]string{"binary": app.Binary},
		}
		res = append(res, dp)
	}
	return res, nil
}
