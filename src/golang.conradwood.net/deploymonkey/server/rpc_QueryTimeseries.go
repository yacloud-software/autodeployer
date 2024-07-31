package main

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/apis/grafanadata"
	"golang.conradwood.net/deploymonkey/db"
	"golang.conradwood.net/go-easyops/errors"
)

func (e *DeployMonkey) QueryTimeseries(req *grafanadata.QueryRequest, srv pb.DeployMonkey_QueryTimeseriesServer) error {
	fmt.Printf("Query timeseries: \"%s\"\n", req.Query)
	var err error
	var dps []*grafanadata.DataPoint
	ctx := srv.Context()
	if req.Query == "deployment_count" {
		dps, err = query_deployment_count(ctx, req)
	} else if req.Query == "version_history" {
		dps, err = query_version_history(ctx, req)
	} else if req.Query == "version_history" {
		dps, err = query_version_history(ctx, req)
	} else if req.Query == "deployments" {
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
	var logs []*pb.DeploymentLog
	var err error
	vb := req.ValueMap["binary"]
	if vb != nil && len(vb.Values) > 0 {
		if len(vb.Values) > 1 {
			return nil, errors.Errorf("cannot yet handle multiple values")
		}
		bin := "%" + vb.Values[0] + "%"
		logs, err = db.DefaultDBDeploymentLog().FromQuery(ctx, "created >= $1 and binary ilike ", req.Start, bin)
	} else {
		logs, err = db.DefaultDBDeploymentLog().FromQuery(ctx, "created >= $1", req.Start)
	}
	if err != nil {
		return nil, err
	}
	var res []*grafanadata.DataPoint

	if req.QueryType == grafanadata.QueryType_ANNOTATIONS {
		for _, log := range logs {
			dp := &grafanadata.DataPoint{
				FieldName:   "text",
				Timestamp:   log.Started,
				StringValue: fmt.Sprintf("%s Build #%d", log.Binary, log.BuildID),
				//Labels:      map[string]string{"build": fmt.Sprintf("%d", app.BuildID)},
			}
			res = append(res, dp)
		}
	} else {
		for _, log := range logs {
			dp := &grafanadata.DataPoint{
				FieldName: "version",
				Timestamp: log.Started,
				Value:     float64(log.BuildID),
				Labels:    map[string]string{"binary": log.Binary},
			}
			res = append(res, dp)
		}
	}

	return res, nil
}

// actually version
func query_version_history(ctx context.Context, req *grafanadata.QueryRequest) ([]*grafanadata.DataPoint, error) {
	apps, err := appdef_store.All(ctx)
	if err != nil {
		return nil, err
	}
	var res []*grafanadata.DataPoint
	var relevant []*pb.ApplicationDefinition
	for _, app := range apps {
		if app.Created < req.Start || app.Created > req.End {
			continue
		}
		relevant = append(relevant, app)
	}

	if req.QueryType == grafanadata.QueryType_ANNOTATIONS {
		for _, app := range relevant {
			dp := &grafanadata.DataPoint{
				FieldName:   "text",
				Timestamp:   app.Created,
				StringValue: fmt.Sprintf("%s Build #%d", app.Binary, app.BuildID),
				//Labels:      map[string]string{"build": fmt.Sprintf("%d", app.BuildID)},
			}
			res = append(res, dp)
		}
	} else {
		for _, app := range relevant {
			dp := &grafanadata.DataPoint{
				FieldName: "version",
				Timestamp: app.Created,
				Value:     float64(app.ID),
				Labels:    map[string]string{"binary": app.Binary},
			}
			res = append(res, dp)
		}
	}
	return res, nil
}
