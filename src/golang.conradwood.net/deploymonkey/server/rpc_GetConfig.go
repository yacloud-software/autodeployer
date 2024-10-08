package main

import (
	"context"
	"golang.conradwood.net/go-easyops/errors"
	//	"errors"
	"fmt"
	common "golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/deploymonkey"
)

// this gets the current CONFIGURATION from the database (not what's actually deployed)
func (depl *DeployMonkey) GetConfig(ctx context.Context, cr *common.Void) (*pb.Config, error) {
	if *debug {
		fmt.Printf("Request to get config\n")
	}
	res := &pb.Config{}
	var err error
	pd, err := depl.GetDeployers(ctx, &common.Void{})
	if err != nil {
		return nil, err
	}
	res.Deployers = pd

	ns, err := depl.GetNameSpaces(ctx, &pb.GetNameSpaceRequest{})
	if err != nil {
		return nil, err
	}
	for _, n := range ns.NameSpaces {
		gns, err := depl.GetGroups(ctx, &pb.GetGroupsRequest{NameSpace: n})
		if err != nil {
			return nil, errors.Errorf("Failed to get groups: %s", err)
		}
		for _, gs := range gns.Groups {
			gc := &pb.GroupConfig{Group: gs}
			gar := pb.GetAppsRequest{NameSpace: gs.NameSpace, GroupName: gs.GroupID}
			gapps, err := depl.GetApplications(ctx, &gar)
			if err != nil {
				return nil, errors.Errorf("Failed to get applications: %s", err)
			}
			gc.Applications = gapps.Applications
			res.GroupConfigs = append(res.GroupConfigs, gc)
		}
	}

	return res, nil
}

func (s *DeployMonkey) GetGroups(ctx context.Context, cr *pb.GetGroupsRequest) (*pb.GetGroupsResponse, error) {
	if cr.NameSpace == "" {
		return nil, errors.Errorf("Namespace required")
	}
	resp := pb.GetGroupsResponse{}

	dbg, err := groupHandler.FindAppGroupByNamespace(ctx, cr.NameSpace)
	//		dbg, err := getGroupFromDatabase(ctx, cr.NameSpace, name)
	if err != nil {
		return nil, err
	}
	gd := pb.GroupDef{
		DeployedVersion: int64(dbg.GetDeployedVersion()),
		PendingVersion:  int64(dbg.GetPendingVersion()),
		GroupID:         fmt.Sprintf("%d", dbg.ID),
		NameSpace:       cr.NameSpace,
	}
	resp.Groups = append(resp.Groups, &gd)

	return &resp, nil
}

func (s *DeployMonkey) GetNameSpaces(ctx context.Context, cr *pb.GetNameSpaceRequest) (*pb.GetNameSpaceResponse, error) {
	resp := pb.GetNameSpaceResponse{}
	n, err := getStringsFromDB("select distinct namespace from appgroup order by namespace asc", "")
	if err != nil {
		return nil, err
	}
	resp.NameSpaces = n
	return &resp, nil
}

func (s *DeployMonkey) GetDeployers(ctx context.Context, cr *common.Void) (*pb.DeployersList, error) {
	autodeployers := GetAllDeployersFromCache()
	dl := &pb.DeployersList{}
	for _, ad := range autodeployers {
		d := &pb.Deployer{Host: ad.IP, Machinegroup: ad.Group}
		dl.Deployers = append(dl.Deployers, d)
	}
	return dl, nil

}

func (s *DeployMonkey) GetApplications(ctx context.Context, cr *pb.GetAppsRequest) (*pb.GetAppsResponse, error) {
	dbg, err := groupHandler.FindAppGroupByNamespace(ctx, cr.NameSpace)
	if err != nil {
		s := fmt.Sprintf("No such group: (%s,%s)\n", cr.NameSpace, cr.GroupName)
		fmt.Println(s)
		return nil, errors.Errorf("%s", s)
	}
	ad, err := loadAppGroupVersion(ctx, dbg.GetDeployedVersion())
	if err != nil {
		s := fmt.Sprintf("GetApplications(): No applications for version %d (%s,%s)", dbg.GetDeployedVersion(), cr.NameSpace, cr.GroupName)
		fmt.Printf("%s: %s\n", s, err)
		return nil, errors.Errorf("%s [%s]", s, err)
	}
	resp := pb.GetAppsResponse{}
	resp.Applications = ad
	return &resp, nil
}
