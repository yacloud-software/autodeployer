package main

import (
	"context"
	"flag"
	"fmt"
	"golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/deployminator"
	"golang.conradwood.net/apis/deploymonkey"
	_ "golang.conradwood.net/deployminator/changes"
	"golang.conradwood.net/deployminator/config"
	"golang.conradwood.net/deployminator/db"
	_ "golang.conradwood.net/deployminator/targets"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/server"
	"golang.conradwood.net/go-easyops/utils"
	"google.golang.org/grpc"
	"os"
)

const (
	NOT_DONE_YET = true
)

var (
	debug = flag.Bool("debug", false, "more logging...")
	port  = flag.Int("port", 4100, "The grpc server port")
)

func main() {
	flag.Parse()
	fmt.Printf("Starting deployminator...\n")
	if NOT_DONE_YET {
		fmt.Printf("This is not completed yet\n")
		os.Exit(0)
	}

	var err error
	db.Open()

	sd := server.NewServerDef()
	sd.Port = *port
	sd.Register = server.Register(
		func(server *grpc.Server) error {
			pb.RegisterDeployminatorServer(server, &Deployminator{})
			return nil
		},
	)
	err = server.ServerStartup(sd)
	utils.Bail("Unable to start server", err)
	os.Exit(0)

}

type Deployminator struct {
}

type tempbuild struct {
	dd *pb.DeploymentDescriptor
	ca []*CompatApp
}

func (d *Deployminator) NewBuildAvailable(ctx context.Context, req *pb.NewBuildRequest) (*common.Void, error) {
	if req.RepositoryID == 0 {
		return nil, errors.InvalidArgs(ctx, "missing repoid", "missing repository id")
	}
	if req.BuildNumber == 0 {
		return nil, errors.InvalidArgs(ctx, "missing buildnumber", "missing buildnumber")
	}
	c, err := config.Parse(req.DeployFile, req.RepositoryID)
	c_apps, err := findOrCreateAppsFromOldConfig(ctx, req.RepositoryID, c)
	if err != nil {
		return nil, err
	}
	if len(c_apps) == 0 {
		return nil, errors.InvalidArgs(ctx, "app not found", "app not found")
	}
	fmt.Printf("Got %d apps:\n", len(c_apps))
	gotnew := false
	instances := make(map[uint64]*tempbuild)
	for _, ca := range c_apps {
		app := ca.app
		fmt.Printf("Found app #%d\n", app.ID)
		dd, isnew, err := get_or_create_deploy_descriptor(ctx, app, req.BuildNumber, req.Branch)
		if err != nil {
			return nil, err
		}
		_, fd := instances[dd.ID]
		if !isnew && !fd {
			fmt.Printf("Is not new (it is deployment descriptor ID %d\n", dd.ID)
			continue
		}

		gotnew = true
		fmt.Printf("DeploymentDescriptor: %d (isnew=%v)\n", dd.ID, isnew)
		tb := instances[dd.ID]
		if tb == nil {
			tb = &tempbuild{dd: dd}
			instances[dd.ID] = tb
		}
		tb.ca = append(tb.ca, ca)

	}
	for _, tb := range instances {
		dd := tb.dd

		err = del_instances(ctx, dd)
		if err != nil {
			return nil, err
		}
		for _, ca := range tb.ca {
			inst, err := add_instances(ctx, dd, ca.machinegroup, ca.instances, false)
			if err != nil {
				return nil, err
			}
			err = set_args(ctx, inst, ca.args)
			if err != nil {
				return nil, err
			}

		}
		err := create_replace_requests(dd)
		if err != nil {
			return nil, err
		}
		//		make_only_deploy(dd)
	}
	fmt.Printf("Changes: %v\n", gotnew)
	return &common.Void{}, err
}
func (d *Deployminator) DeployVersion(ctx context.Context, req *pb.DeployRequest) (*common.Void, error) {
	return nil, errors.NotImplemented(ctx, "not yet")
}
func (d *Deployminator) ListDeployments(ctx context.Context, req *common.Void) (*deploymonkey.DeploymentList, error) {
	return nil, errors.NotImplemented(ctx, "not yet")
}
func (d *Deployminator) UndeployVersion(ctx context.Context, req *pb.UndeployRequest) (*common.Void, error) {
	return nil, errors.NotImplemented(ctx, "not yet")
}
func (d *Deployminator) AutodeployerStartup(ctx context.Context, req *common.Void) (*common.Void, error) {
	return nil, errors.NotImplemented(ctx, "not yet")
}
func (d *Deployminator) AutodeployerShutdown(ctx context.Context, req *common.Void) (*common.Void, error) {
	return nil, errors.NotImplemented(ctx, "not yet")
}

func del_instances(ctx context.Context, dd *pb.DeploymentDescriptor) error {
	_, err := db.Psql.ExecContext(ctx, "del_instances", "delete from "+db.Instancedb.Tablename()+" where deploymentid=$1", dd.ID)
	return err

}

func add_instances(ctx context.Context, dd *pb.DeploymentDescriptor, machinegroup string, instances uint32, permachine bool) (*pb.InstanceDef, error) {
	id := &pb.InstanceDef{
		DeploymentID:              dd,
		MachineGroup:              machinegroup,
		Instances:                 instances,
		InstanceCountIsPerMachine: permachine,
	}
	fmt.Printf("deploymentid %d: Adding %d instances on machine group %s\n", id.DeploymentID.ID, id.Instances, id.MachineGroup)
	_, err := db.Instancedb.Save(ctx, id)
	if err != nil {
		return nil, err
	}
	return id, nil
}

func make_only_deploy(req *pb.DeploymentDescriptor) error {
	ctx := authremote.Context()
	_, err := db.Psql.ExecContext(ctx, "undeployme_all_for_app", "update "+db.Descriptordb.Tablename()+" set deployme = false where application=$1 and branch=$2", req.Application.ID, req.Branch)
	if err != nil {
		return err
	}
	req.DeployMe = true
	err = db.Descriptordb.Update(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

// given a new deployment descriptor removes all previous versions and creates change requests for each
func create_replace_requests(req *pb.DeploymentDescriptor) error {
	ctx := authremote.Context()
	dids, err := db.Descriptordb.FromQuery(ctx, "application=$1 and branch=$2 and id != $3 and deployme = true", req.Application.ID, req.Branch, req.ID)
	if err != nil {
		return err
	}
	for _, did := range dids {
		rr := &pb.ReplaceRequest{OldDeployment: did, NewDeployment: req}
		_, err = db.Replacedb.Save(ctx, rr)
		if err != nil {
			return err
		}
	}
	err = make_only_deploy(req)
	if err != nil {
		return err
	}
	return nil
}
