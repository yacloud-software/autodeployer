package db

import (
	"context"
	pb "golang.conradwood.net/apis/deployminator"
	"golang.conradwood.net/go-easyops/sql"
	"golang.conradwood.net/go-easyops/utils"
)

var (
	Psql         *sql.DB
	Descriptordb *DBDeploymentDescriptor
	Appdb        *DBApplication
	Argsdb       *DBArgument
	Instancedb   *DBInstanceDef
	Replacedb    *DBReplaceRequest
)

func Open() {
	var err error
	Psql, err = sql.Open()
	utils.Bail("failed to open db", err)
	Descriptordb = NewDBDeploymentDescriptor(Psql)
	Appdb = NewDBApplication(Psql)
	Argsdb = NewDBArgument(Psql)
	Instancedb = NewDBInstanceDef(Psql)
	Replacedb = NewDBReplaceRequest(Psql)

}

type Instarg struct {
	Instance *pb.InstanceDef
	args     []*pb.Argument
}
type FullDD struct {
	DeploymentDescriptor *pb.DeploymentDescriptor
	Instances            []*Instarg
}

func FetchFull(ctx context.Context, dd *pb.DeploymentDescriptor) (*FullDD, error) {
	var err error
	res := &FullDD{DeploymentDescriptor: dd}
	res.DeploymentDescriptor.Application, err = Appdb.ByID(ctx, res.DeploymentDescriptor.Application.ID)
	if err != nil {
		return nil, err
	}

	instances, err := Instancedb.ByDeploymentID(ctx, dd.ID)
	if err != nil {
		return nil, err
	}
	for _, i := range instances {
		args, err := Argsdb.ByInstanceDef(ctx, dd.ID)
		if err != nil {
			return nil, err
		}
		ia := &Instarg{Instance: i, args: args}
		res.Instances = append(res.Instances, ia)
	}
	return res, nil
}
