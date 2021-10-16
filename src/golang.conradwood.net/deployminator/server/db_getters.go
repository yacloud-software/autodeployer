package main

import (
	"context"
	pb "golang.conradwood.net/apis/deployminator"
	"golang.conradwood.net/deployminator/db"
)

func set_args(ctx context.Context, dd *pb.InstanceDef, args []string) error {
	_, err := db.Psql.ExecContext(ctx, "del_args", "delete from "+db.Argsdb.Tablename()+" where instancedef = $1", dd.ID)
	if err != nil {
		return err
	}
	for _, a := range args {
		a := &pb.Argument{InstanceDef: dd, Argument: a}
		_, err = db.Argsdb.Save(ctx, a)
		if err != nil {
			return err
		}
	}
	return nil
}

// returns deployment descriptor. true if newly created, false if not
func get_or_create_deploy_descriptor(ctx context.Context, app *pb.Application, build uint64, branch string) (*pb.DeploymentDescriptor, bool, error) {
	dds, err := db.Descriptordb.FromQuery(ctx, "application = $1 and buildnumber = $2 and branch = $3", app.ID, build, branch)
	if err != nil {
		return nil, false, err
	}
	if len(dds) != 0 {
		return dds[0], false, nil
	}
	dd := &pb.DeploymentDescriptor{
		Application: app,
		BuildNumber: build,
		Branch:      branch,
	}
	_, err = db.Descriptordb.Save(ctx, dd)
	return dd, true, nil
}
func findOrCreateApp(ctx context.Context, repoid uint64, binary, url string) (*pb.Application, error) {
	apps, err := db.Appdb.FromQuery(ctx, "repositoryid = $1 and r_binary=$2 and downloadurl=$3", repoid, binary, url)
	if err != nil {
		return nil, err
	}
	if len(apps) != 0 {
		return apps[0], nil
	}

	ap := &pb.Application{
		RepositoryID: repoid,
		Binary:       binary,
		DownloadURL:  url,
	}
	_, err = db.Appdb.Save(ctx, ap)
	if err != nil {
		return nil, err
	}
	return ap, nil
}
