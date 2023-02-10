package main

import (
	"context"
	"fmt"
	"golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/deploymonkey"
)

func (d *DeployMonkey) DeleteVersion(ctx context.Context, req *pb.DelVersionRequest) (*common.Void, error) {
	fmt.Printf("Deleting version %d\n", req.Version)
	apps, err := loadAppGroupVersion(ctx, int(req.Version))
	if err != nil {
		return nil, err
	}
	uar := &pb.UndeployApplicationRequest{ID: int64(req.Version)}
	_, err = d.UndeployApplication(ctx, uar)
	if err != nil {
		return nil, err
	}
	for _, a := range apps {
		err = appdef_store.DeleteByID(ctx, a.ID)
		if err != nil {
			return nil, err
		}
		fmt.Printf("App: %#v\n", a)
	}
	return &common.Void{}, nil
}
