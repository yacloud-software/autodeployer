package main

import (
	"context"
	"fmt"

	"golang.conradwood.net/apis/common"
	dm "golang.conradwood.net/apis/deploymonkey"
	dc "golang.conradwood.net/deploymonkey/common"
	"golang.conradwood.net/deploymonkey/useroverride"
	"golang.conradwood.net/go-easyops/errors"
)

func (depl *DeployMonkey) NewBuildAvailable(ctx context.Context, req *dm.NewBuildAvailableRequest) (*common.Void, error) {
	fd, err := dc.ParseConfig(req.DeployYaml, 0)
	if err != nil {
		return nil, errors.Errorf("parser failed: %w", err)
	}
	for _, group := range fd.Groups {
		// add stuff to group (which isn't in deploy.yaml, but in metadata instead
		for _, app := range group.Applications {
			app.ArtefactID = req.ArtefactID
			if req.RepositoryID != 0 && app.RepositoryID == 0 {
				app.RepositoryID = req.RepositoryID
			}
			app.BuildID = req.BuildID
			useroverride.MarkAsDeployed(app)
		}
	}
	// check each app is complete
	err = CheckCompleteConfigFile(fd)
	if err != nil {
		return nil, err
	}

	for _, group := range fd.Groups {
		fmt.Printf("Creating new group for build %d,namespace=%s,groupid=%s (ArtefactID %d)\n", req.BuildID, group.Namespace, group.GroupID, req.ArtefactID)
		gv, err := groupHandler.CreateGroupVersion(ctx, group)
		if err != nil {
			return nil, err
		}
		fmt.Printf("New Group Version: %v\n", gv)
		dr := &dm.DeployRequest{VersionID: fmt.Sprintf("%d", gv.ID)}
		_, err = depl.DeployVersion(ctx, dr)
		if err != nil {
			return nil, err
		}
	}
	return &common.Void{}, nil
}
func CheckCompleteConfigFile(fd *dc.FileDef) error {
	for _, group := range fd.Groups {
		for _, app := range group.Applications {
			err := dc.CheckAppComplete(app)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
