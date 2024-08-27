package main

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/go-easyops/errors"
)

func ConvertOldApp(ctx context.Context, id uint64) (*pb.ApplicationDefinition, error) {
	rows, err := dbcon.QueryContext(ctx, "queryname", "SELECT appdef.id,sourceurl,downloaduser,downloadpw,executable,buildid,instances,mgroup,deploytype,critical,alwayson,statictargetdir,ispublic,java from appdef, lnk_app_grp where appdef.id = lnk_app_grp.app_id and app_id = $1", id)
	if err != nil {
		fmt.Printf("oldapp: Failed to get app with id %d:%s\n", id, err)
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, errors.Errorf("oldapp no app with id %d", id)
	}
	res := &pb.ApplicationDefinition{}

	err = rows.Scan(&res.ID, &res.DownloadURL, &res.DownloadUser, &res.DownloadPassword,
		&res.Binary, &res.BuildID, &res.Instances, &res.Machines, &res.DeployType, &res.Critical, &res.AlwaysOn, &res.StaticTargetDir, &res.Public, &res.Java)
	if err != nil {
		return nil, errors.Errorf("oldapp (%s)", err)
	}

	return res, nil
}
