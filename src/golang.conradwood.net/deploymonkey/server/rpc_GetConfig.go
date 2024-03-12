package main

import (
	"context"
	"errors"
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
			return nil, fmt.Errorf("Failed to get groups: %s", err)
		}
		for _, gs := range gns.Groups {
			gc := &pb.GroupConfig{Group: gs}
			gar := pb.GetAppsRequest{NameSpace: gs.NameSpace, GroupName: gs.GroupID}
			gapps, err := depl.GetApplications(ctx, &gar)
			if err != nil {
				return nil, fmt.Errorf("Failed to get applications: %s", err)
			}
			gc.Applications = gapps.Applications
			res.GroupConfigs = append(res.GroupConfigs, gc)
		}
	}

	return res, nil
}

func (s *DeployMonkey) GetGroups(ctx context.Context, cr *pb.GetGroupsRequest) (*pb.GetGroupsResponse, error) {
	if cr.NameSpace == "" {
		return nil, errors.New("Namespace required")
	}
	resp := pb.GetGroupsResponse{}
	n, err := getStringsFromDB("select groupname from appgroup where namespace = $1 order by groupname asc", cr.NameSpace)
	if err != nil {
		return nil, err
	}
	for _, name := range n {
		dbg, err := getGroupFromDatabase(ctx, cr.NameSpace, name)
		if err != nil {
			return nil, err
		}
		gd := pb.GroupDef{DeployedVersion: int64(dbg.DeployedVersion),
			PendingVersion: int64(dbg.PendingVersion),
			GroupID:        dbg.groupDef.GroupID,
			NameSpace:      dbg.groupDef.Namespace,
		}
		resp.Groups = append(resp.Groups, &gd)

	}

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
	dbg, err := getGroupFromDatabase(ctx, cr.NameSpace, cr.GroupName)
	if err != nil {
		s := fmt.Sprintf("No such group: (%s,%s)\n", cr.NameSpace, cr.GroupName)
		fmt.Println(s)
		return nil, errors.New(s)
	}
	ad, err := loadAppGroupVersion(ctx, dbg.DeployedVersion)
	if err != nil {
		s := fmt.Sprintf("GetApplications(): No applications for version %d (%s,%s)", dbg.DeployedVersion, cr.NameSpace, cr.GroupName)
		fmt.Printf("%s: %s\n", s, err)
		return nil, fmt.Errorf("%s [%s]", s, err)
	}
	resp := pb.GetAppsResponse{}
	resp.Applications = ad
	return &resp, nil
}
