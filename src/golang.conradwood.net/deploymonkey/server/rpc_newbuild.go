package main

import (
	"context"
	"fmt"
	"golang.conradwood.net/apis/common"
	dm "golang.conradwood.net/apis/deploymonkey"
	dc "golang.conradwood.net/deploymonkey/common"
)

func (depl *DeployMonkey) NewBuildAvailable(ctx context.Context, req *dm.NewBuildAvailableRequest) (*common.Void, error) {
	fd, err := dc.ParseConfig(req.DeployYaml, 0)
	if err != nil {
		return nil, fmt.Errorf("parser failed: %w", err)
	}

	for _, group := range fd.Groups {
		// add stuff to group (which isn't in deploy.yaml, but in metadata instead
		for _, app := range group.Applications {
			app.ArtefactID = req.ArtefactID
			if req.RepositoryID != 0 && app.RepositoryID == 0 {
				app.RepositoryID = req.RepositoryID
			}
		}
		// save new submitted stuff
		resp, err := depl.DefineGroup(ctx, group)
		if err != nil {
			fmt.Printf("Failed to define group: %s\n", err)
			return nil, fmt.Errorf("failed to define group: %w", err)
		}
		if resp.Result != dm.GroupResponseStatus_CHANGEACCEPTED {
			fmt.Printf("Response to deploy: %s - skipping\n", resp.Result)
			continue
		}
		dr := dm.DeployRequest{VersionID: resp.VersionID}
		dresp, err := depl.DeployVersion(ctx, &dr)
		if err != nil {
			fmt.Printf("Failed to deploy version %s: %s\n", resp.VersionID, err)
			return nil, err
		}
		fmt.Printf("Deploy response: %v\n", dresp)
	}
	return &common.Void{}, nil
}
