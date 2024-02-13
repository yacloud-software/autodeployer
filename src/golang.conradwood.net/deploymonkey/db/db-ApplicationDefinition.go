package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBApplicationDefinition
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence applicationdefinition_seq;

Main Table:

 CREATE TABLE applicationdefinition (id integer primary key default nextval('applicationdefinition_seq'),downloadurl text not null  ,downloaduser text not null  ,downloadpassword text not null  ,r_binary text not null  ,buildid bigint not null  ,instances integer not null  ,deploymentid text not null  ,machines text not null  ,deploytype text not null  ,critical boolean not null  ,alwayson boolean not null  ,statictargetdir text not null  ,r_public boolean not null  ,java boolean not null  ,repositoryid bigint not null  ,asroot boolean not null  ,container bigint not null  references containerdef (id) on delete cascade  ,discardlog boolean not null  ,artefactid bigint not null  );

Alter statements:
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS downloadurl text not null default '';
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS downloaduser text not null default '';
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS downloadpassword text not null default '';
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS r_binary text not null default '';
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS buildid bigint not null default 0;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS instances integer not null default 0;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS deploymentid text not null default '';
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS machines text not null default '';
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS deploytype text not null default '';
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS critical boolean not null default false;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS alwayson boolean not null default false;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS statictargetdir text not null default '';
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS r_public boolean not null default false;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS java boolean not null default false;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS asroot boolean not null default false;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS container bigint not null references containerdef (id) on delete cascade  default 0;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS discardlog boolean not null default false;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS artefactid bigint not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE applicationdefinition_archive (id integer unique not null,downloadurl text not null,downloaduser text not null,downloadpassword text not null,r_binary text not null,buildid bigint not null,instances integer not null,deploymentid text not null,machines text not null,deploytype text not null,critical boolean not null,alwayson boolean not null,statictargetdir text not null,r_public boolean not null,java boolean not null,repositoryid bigint not null,asroot boolean not null,container bigint not null,discardlog boolean not null,artefactid bigint not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/go-easyops/sql"
	"os"
)

var (
	default_def_DBApplicationDefinition *DBApplicationDefinition
)

type DBApplicationDefinition struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBApplicationDefinition() *DBApplicationDefinition {
	if default_def_DBApplicationDefinition != nil {
		return default_def_DBApplicationDefinition
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBApplicationDefinition(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBApplicationDefinition = res
	return res
}
func NewDBApplicationDefinition(db *sql.DB) *DBApplicationDefinition {
	foo := DBApplicationDefinition{DB: db}
	foo.SQLTablename = "applicationdefinition"
	foo.SQLArchivetablename = "applicationdefinition_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBApplicationDefinition) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBApplicationDefinition", "insert into "+a.SQLArchivetablename+" (id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20) ", p.ID, p.DownloadURL, p.DownloadUser, p.DownloadPassword, p.Binary, p.BuildID, p.Instances, p.DeploymentID, p.Machines, p.DeployType, p.Critical, p.AlwaysOn, p.StaticTargetDir, p.Public, p.Java, p.RepositoryID, p.AsRoot, p.Container.ID, p.DiscardLog, p.ArtefactID)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBApplicationDefinition) Save(ctx context.Context, p *savepb.ApplicationDefinition) (uint64, error) {
	qn := "DBApplicationDefinition_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19) returning id", p.DownloadURL, p.DownloadUser, p.DownloadPassword, p.Binary, p.BuildID, p.Instances, p.DeploymentID, p.Machines, p.DeployType, p.Critical, p.AlwaysOn, p.StaticTargetDir, p.Public, p.Java, p.RepositoryID, p.AsRoot, p.Container.ID, p.DiscardLog, p.ArtefactID)
	if e != nil {
		return 0, a.Error(ctx, qn, e)
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, a.Error(ctx, qn, fmt.Errorf("No rows after insert"))
	}
	var id uint64
	e = rows.Scan(&id)
	if e != nil {
		return 0, a.Error(ctx, qn, fmt.Errorf("failed to scan id after insert: %s", e))
	}
	p.ID = id
	return id, nil
}

// Save using the ID specified
func (a *DBApplicationDefinition) SaveWithID(ctx context.Context, p *savepb.ApplicationDefinition) error {
	qn := "insert_DBApplicationDefinition"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20) ", p.ID, p.DownloadURL, p.DownloadUser, p.DownloadPassword, p.Binary, p.BuildID, p.Instances, p.DeploymentID, p.Machines, p.DeployType, p.Critical, p.AlwaysOn, p.StaticTargetDir, p.Public, p.Java, p.RepositoryID, p.AsRoot, p.Container.ID, p.DiscardLog, p.ArtefactID)
	return a.Error(ctx, qn, e)
}

func (a *DBApplicationDefinition) Update(ctx context.Context, p *savepb.ApplicationDefinition) error {
	qn := "DBApplicationDefinition_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set downloadurl=$1, downloaduser=$2, downloadpassword=$3, r_binary=$4, buildid=$5, instances=$6, deploymentid=$7, machines=$8, deploytype=$9, critical=$10, alwayson=$11, statictargetdir=$12, r_public=$13, java=$14, repositoryid=$15, asroot=$16, container=$17, discardlog=$18, artefactid=$19 where id = $20", p.DownloadURL, p.DownloadUser, p.DownloadPassword, p.Binary, p.BuildID, p.Instances, p.DeploymentID, p.Machines, p.DeployType, p.Critical, p.AlwaysOn, p.StaticTargetDir, p.Public, p.Java, p.RepositoryID, p.AsRoot, p.Container.ID, p.DiscardLog, p.ArtefactID, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBApplicationDefinition) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBApplicationDefinition_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBApplicationDefinition) ByID(ctx context.Context, p uint64) (*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No ApplicationDefinition with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) ApplicationDefinition with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBApplicationDefinition) TryByID(ctx context.Context, p uint64) (*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("TryByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) ApplicationDefinition with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBApplicationDefinition) All(ctx context.Context) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" order by id")
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("All: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, fmt.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBApplicationDefinition" rows with matching DownloadURL
func (a *DBApplicationDefinition) ByDownloadURL(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDownloadURL"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where downloadurl = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadURL: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadURL: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeDownloadURL(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeDownloadURL"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where downloadurl ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadURL: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadURL: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching DownloadUser
func (a *DBApplicationDefinition) ByDownloadUser(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDownloadUser"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where downloaduser = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadUser: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadUser: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeDownloadUser(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeDownloadUser"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where downloaduser ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadUser: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadUser: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching DownloadPassword
func (a *DBApplicationDefinition) ByDownloadPassword(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDownloadPassword"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where downloadpassword = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadPassword: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadPassword: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeDownloadPassword(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeDownloadPassword"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where downloadpassword ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadPassword: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadPassword: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Binary
func (a *DBApplicationDefinition) ByBinary(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByBinary"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where r_binary = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBinary: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBinary: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeBinary(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeBinary"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where r_binary ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBinary: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBinary: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching BuildID
func (a *DBApplicationDefinition) ByBuildID(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByBuildID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where buildid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBuildID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBuildID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeBuildID(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeBuildID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where buildid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBuildID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBuildID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Instances
func (a *DBApplicationDefinition) ByInstances(ctx context.Context, p uint32) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByInstances"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where instances = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInstances: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInstances: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeInstances(ctx context.Context, p uint32) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeInstances"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where instances ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInstances: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInstances: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching DeploymentID
func (a *DBApplicationDefinition) ByDeploymentID(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDeploymentID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where deploymentid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeploymentID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeploymentID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeDeploymentID(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeDeploymentID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where deploymentid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeploymentID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeploymentID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Machines
func (a *DBApplicationDefinition) ByMachines(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByMachines"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where machines = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMachines: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMachines: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeMachines(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeMachines"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where machines ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMachines: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMachines: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching DeployType
func (a *DBApplicationDefinition) ByDeployType(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDeployType"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where deploytype = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployType: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployType: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeDeployType(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeDeployType"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where deploytype ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployType: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployType: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Critical
func (a *DBApplicationDefinition) ByCritical(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByCritical"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where critical = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCritical: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCritical: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeCritical(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeCritical"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where critical ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCritical: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCritical: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching AlwaysOn
func (a *DBApplicationDefinition) ByAlwaysOn(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByAlwaysOn"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where alwayson = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAlwaysOn: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAlwaysOn: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeAlwaysOn(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeAlwaysOn"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where alwayson ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAlwaysOn: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAlwaysOn: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching StaticTargetDir
func (a *DBApplicationDefinition) ByStaticTargetDir(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByStaticTargetDir"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where statictargetdir = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByStaticTargetDir: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByStaticTargetDir: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeStaticTargetDir(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeStaticTargetDir"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where statictargetdir ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByStaticTargetDir: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByStaticTargetDir: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Public
func (a *DBApplicationDefinition) ByPublic(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByPublic"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where r_public = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPublic: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPublic: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikePublic(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikePublic"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where r_public ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPublic: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPublic: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Java
func (a *DBApplicationDefinition) ByJava(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByJava"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where java = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByJava: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByJava: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeJava(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeJava"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where java ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByJava: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByJava: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching RepositoryID
func (a *DBApplicationDefinition) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching AsRoot
func (a *DBApplicationDefinition) ByAsRoot(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByAsRoot"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where asroot = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAsRoot: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAsRoot: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeAsRoot(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeAsRoot"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where asroot ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAsRoot: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAsRoot: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Container
func (a *DBApplicationDefinition) ByContainer(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByContainer"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where container = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByContainer: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByContainer: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeContainer(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeContainer"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where container ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByContainer: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByContainer: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching DiscardLog
func (a *DBApplicationDefinition) ByDiscardLog(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDiscardLog"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where discardlog = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDiscardLog: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDiscardLog: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeDiscardLog(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeDiscardLog"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where discardlog ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDiscardLog: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDiscardLog: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching ArtefactID
func (a *DBApplicationDefinition) ByArtefactID(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByArtefactID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where artefactid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByArtefactID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByArtefactID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeArtefactID(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeArtefactID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid from "+a.SQLTablename+" where artefactid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByArtefactID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByArtefactID: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBApplicationDefinition) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.ApplicationDefinition, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBApplicationDefinition) Tablename() string {
	return a.SQLTablename
}

func (a *DBApplicationDefinition) SelectCols() string {
	return "id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid"
}
func (a *DBApplicationDefinition) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".downloadurl, " + a.SQLTablename + ".downloaduser, " + a.SQLTablename + ".downloadpassword, " + a.SQLTablename + ".r_binary, " + a.SQLTablename + ".buildid, " + a.SQLTablename + ".instances, " + a.SQLTablename + ".deploymentid, " + a.SQLTablename + ".machines, " + a.SQLTablename + ".deploytype, " + a.SQLTablename + ".critical, " + a.SQLTablename + ".alwayson, " + a.SQLTablename + ".statictargetdir, " + a.SQLTablename + ".r_public, " + a.SQLTablename + ".java, " + a.SQLTablename + ".repositoryid, " + a.SQLTablename + ".asroot, " + a.SQLTablename + ".container, " + a.SQLTablename + ".discardlog, " + a.SQLTablename + ".artefactid"
}

func (a *DBApplicationDefinition) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.ApplicationDefinition, error) {
	var res []*savepb.ApplicationDefinition
	for rows.Next() {
		foo := savepb.ApplicationDefinition{Limits: &savepb.Limits{}, Container: &savepb.ContainerDef{}}
		err := rows.Scan(&foo.ID, &foo.DownloadURL, &foo.DownloadUser, &foo.DownloadPassword, &foo.Binary, &foo.BuildID, &foo.Instances, &foo.DeploymentID, &foo.Machines, &foo.DeployType, &foo.Critical, &foo.AlwaysOn, &foo.StaticTargetDir, &foo.Public, &foo.Java, &foo.RepositoryID, &foo.AsRoot, &foo.Container.ID, &foo.DiscardLog, &foo.ArtefactID)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}
func (a *DBApplicationDefinition) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.ApplicationDefinition, error) {
	var res []*savepb.ApplicationDefinition
	for rows.Next() {
		// SCANNER:
		foo := &savepb.ApplicationDefinition{}
		// create the non-nullable pointers
		foo.Container = &savepb.ContainerDef{} // non-nullable
		// create variables for scan results
		scanTarget_0 := &foo.DownloadURL
		scanTarget_1 := &foo.DownloadUser
		scanTarget_2 := &foo.DownloadPassword
		scanTarget_3 := &foo.Binary
		scanTarget_4 := &foo.BuildID
		scanTarget_5 := &foo.Instances
		scanTarget_6 := &foo.DeploymentID
		scanTarget_7 := &foo.Machines
		scanTarget_8 := &foo.DeployType
		scanTarget_9 := &foo.ID
		scanTarget_10 := &foo.Critical
		scanTarget_11 := &foo.AlwaysOn
		scanTarget_12 := &foo.StaticTargetDir
		scanTarget_13 := &foo.Public
		scanTarget_14 := &foo.Java
		scanTarget_15 := &foo.RepositoryID
		scanTarget_16 := &foo.AsRoot
		scanTarget_17 := &foo.Container.ID
		scanTarget_18 := &foo.DiscardLog
		scanTarget_19 := &foo.ArtefactID
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4, scanTarget_5, scanTarget_6, scanTarget_7, scanTarget_8, scanTarget_9, scanTarget_10, scanTarget_11, scanTarget_12, scanTarget_13, scanTarget_14, scanTarget_15, scanTarget_16, scanTarget_17, scanTarget_18, scanTarget_19)
		// END SCANNER

		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, foo)
	}
	return res, nil
}

/**********************************************************************
* Helper to create table and columns
**********************************************************************/
func (a *DBApplicationDefinition) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),downloadurl text not null ,downloaduser text not null ,downloadpassword text not null ,r_binary text not null ,buildid bigint not null ,instances integer not null ,deploymentid text not null ,machines text not null ,deploytype text not null ,critical boolean not null ,alwayson boolean not null ,statictargetdir text not null ,r_public boolean not null ,java boolean not null ,repositoryid bigint not null ,asroot boolean not null ,container bigint not null ,discardlog boolean not null ,artefactid bigint not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),downloadurl text not null ,downloaduser text not null ,downloadpassword text not null ,r_binary text not null ,buildid bigint not null ,instances integer not null ,deploymentid text not null ,machines text not null ,deploytype text not null ,critical boolean not null ,alwayson boolean not null ,statictargetdir text not null ,r_public boolean not null ,java boolean not null ,repositoryid bigint not null ,asroot boolean not null ,container bigint not null ,discardlog boolean not null ,artefactid bigint not null );`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS downloadurl text not null default '';`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS downloaduser text not null default '';`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS downloadpassword text not null default '';`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS r_binary text not null default '';`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS buildid bigint not null default 0;`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS instances integer not null default 0;`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS deploymentid text not null default '';`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS machines text not null default '';`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS deploytype text not null default '';`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS critical boolean not null default false;`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS alwayson boolean not null default false;`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS statictargetdir text not null default '';`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS r_public boolean not null default false;`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS java boolean not null default false;`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS asroot boolean not null default false;`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS container bigint not null default 0;`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS discardlog boolean not null default false;`,
		`ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS artefactid bigint not null default 0;`,

		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS downloadurl text not null  default '';`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS downloaduser text not null  default '';`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS downloadpassword text not null  default '';`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS r_binary text not null  default '';`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS buildid bigint not null  default 0;`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS instances integer not null  default 0;`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS deploymentid text not null  default '';`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS machines text not null  default '';`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS deploytype text not null  default '';`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS critical boolean not null  default false;`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS alwayson boolean not null  default false;`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS statictargetdir text not null  default '';`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS r_public boolean not null  default false;`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS java boolean not null  default false;`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS repositoryid bigint not null  default 0;`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS asroot boolean not null  default false;`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS container bigint not null  default 0;`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS discardlog boolean not null  default false;`,
		`ALTER TABLE applicationdefinition_archive ADD COLUMN IF NOT EXISTS artefactid bigint not null  default 0;`,
	}

	for i, c := range csql {
		_, e := a.DB.ExecContext(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
		if e != nil {
			return e
		}
	}

	// these are optional, expected to fail
	csql = []string{
		// Indices:

		// Foreign keys:
		`ALTER TABLE applicationdefinition add constraint mkdb_fk_applicationdefinition_container_containerdefid FOREIGN KEY (container) references containerdef (id) on delete cascade ;`,
	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBApplicationDefinition) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

