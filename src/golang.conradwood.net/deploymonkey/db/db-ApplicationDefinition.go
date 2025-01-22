package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBApplicationDefinition
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence applicationdefinition_seq;

Main Table:

 CREATE TABLE applicationdefinition (id integer primary key default nextval('applicationdefinition_seq'),downloadurl text not null  ,downloaduser text not null  ,downloadpassword text not null  ,r_binary text not null  ,buildid bigint not null  ,instances integer not null  ,deploymentid text not null  ,machines text not null  ,deploytype text not null  ,critical boolean not null  ,alwayson boolean not null  ,statictargetdir text not null  ,r_public boolean not null  ,java boolean not null  ,repositoryid bigint not null  ,asroot boolean not null  ,container bigint   references containerdef (id) on delete cascade  references INVALID REFERENCE: "true" (INVALID REFERENCE: "true") on delete cascade  ,discardlog boolean not null  ,artefactid bigint not null  ,created integer not null  ,instancesmeansperautodeployer boolean not null  );

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
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS container bigint  references containerdef (id) on delete cascade  references INVALID REFERENCE: "true" (INVALID REFERENCE: "true") on delete cascade  default 0;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS discardlog boolean not null default false;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS artefactid bigint not null default 0;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS created integer not null default 0;
ALTER TABLE applicationdefinition ADD COLUMN IF NOT EXISTS instancesmeansperautodeployer boolean not null default false;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE applicationdefinition_archive (id integer unique not null,downloadurl text not null,downloaduser text not null,downloadpassword text not null,r_binary text not null,buildid bigint not null,instances integer not null,deploymentid text not null,machines text not null,deploytype text not null,critical boolean not null,alwayson boolean not null,statictargetdir text not null,r_public boolean not null,java boolean not null,repositoryid bigint not null,asroot boolean not null,container bigint ,discardlog boolean not null,artefactid bigint not null,created integer not null,instancesmeansperautodeployer boolean not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/sql"
	"os"
	"sync"
)

var (
	default_def_DBApplicationDefinition *DBApplicationDefinition
)

type DBApplicationDefinition struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
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

func (a *DBApplicationDefinition) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBApplicationDefinition) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

// archive. It is NOT transactionally save.
func (a *DBApplicationDefinition) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBApplicationDefinition", "insert into "+a.SQLArchivetablename+" (id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid, created, instancesmeansperautodeployer) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22) ", p.ID, p.DownloadURL, p.DownloadUser, p.DownloadPassword, p.Binary, p.BuildID, p.Instances, p.DeploymentID, p.Machines, p.DeployType, p.Critical, p.AlwaysOn, p.StaticTargetDir, p.Public, p.Java, p.RepositoryID, p.AsRoot, p.Container.ID, p.DiscardLog, p.ArtefactID, p.Created, p.InstancesMeansPerAutodeployer)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBApplicationDefinition) buildSaveMap(ctx context.Context, p *savepb.ApplicationDefinition) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["downloadurl"] = a.get_col_from_proto(p, "downloadurl")
	res["downloaduser"] = a.get_col_from_proto(p, "downloaduser")
	res["downloadpassword"] = a.get_col_from_proto(p, "downloadpassword")
	res["r_binary"] = a.get_col_from_proto(p, "r_binary")
	res["buildid"] = a.get_col_from_proto(p, "buildid")
	res["instances"] = a.get_col_from_proto(p, "instances")
	res["deploymentid"] = a.get_col_from_proto(p, "deploymentid")
	res["machines"] = a.get_col_from_proto(p, "machines")
	res["deploytype"] = a.get_col_from_proto(p, "deploytype")
	res["critical"] = a.get_col_from_proto(p, "critical")
	res["alwayson"] = a.get_col_from_proto(p, "alwayson")
	res["statictargetdir"] = a.get_col_from_proto(p, "statictargetdir")
	res["r_public"] = a.get_col_from_proto(p, "r_public")
	res["java"] = a.get_col_from_proto(p, "java")
	res["repositoryid"] = a.get_col_from_proto(p, "repositoryid")
	res["asroot"] = a.get_col_from_proto(p, "asroot")
	res["container"] = a.get_col_from_proto(p, "container")
	res["discardlog"] = a.get_col_from_proto(p, "discardlog")
	res["artefactid"] = a.get_col_from_proto(p, "artefactid")
	res["created"] = a.get_col_from_proto(p, "created")
	res["instancesmeansperautodeployer"] = a.get_col_from_proto(p, "instancesmeansperautodeployer")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBApplicationDefinition) Save(ctx context.Context, p *savepb.ApplicationDefinition) (uint64, error) {
	qn := "save_DBApplicationDefinition"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBApplicationDefinition) SaveWithID(ctx context.Context, p *savepb.ApplicationDefinition) error {
	qn := "insert_DBApplicationDefinition"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBApplicationDefinition) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.ApplicationDefinition) (uint64, error) {
	// Save (and use database default ID generation)

	var rows *gosql.Rows
	var e error

	q_cols := ""
	q_valnames := ""
	q_vals := make([]interface{}, 0)
	deli := ""
	i := 0
	// build the 2 parts of the query (column names and value names) as well as the values themselves
	for colname, val := range smap {
		q_cols = q_cols + deli + colname
		i++
		q_valnames = q_valnames + deli + fmt.Sprintf("$%d", i)
		q_vals = append(q_vals, val)
		deli = ","
	}
	rows, e = a.DB.QueryContext(ctx, queryname, "insert into "+a.SQLTablename+" ("+q_cols+") values ("+q_valnames+") returning id", q_vals...)
	if e != nil {
		return 0, a.Error(ctx, queryname, e)
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, a.Error(ctx, queryname, errors.Errorf("No rows after insert"))
	}
	var id uint64
	e = rows.Scan(&id)
	if e != nil {
		return 0, a.Error(ctx, queryname, errors.Errorf("failed to scan id after insert: %s", e))
	}
	p.ID = id
	return id, nil
}

func (a *DBApplicationDefinition) Update(ctx context.Context, p *savepb.ApplicationDefinition) error {
	qn := "DBApplicationDefinition_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set downloadurl=$1, downloaduser=$2, downloadpassword=$3, r_binary=$4, buildid=$5, instances=$6, deploymentid=$7, machines=$8, deploytype=$9, critical=$10, alwayson=$11, statictargetdir=$12, r_public=$13, java=$14, repositoryid=$15, asroot=$16, container=$17, discardlog=$18, artefactid=$19, created=$20, instancesmeansperautodeployer=$21 where id = $22", a.get_DownloadURL(p), a.get_DownloadUser(p), a.get_DownloadPassword(p), a.get_Binary(p), a.get_BuildID(p), a.get_Instances(p), a.get_DeploymentID(p), a.get_Machines(p), a.get_DeployType(p), a.get_Critical(p), a.get_AlwaysOn(p), a.get_StaticTargetDir(p), a.get_Public(p), a.get_Java(p), a.get_RepositoryID(p), a.get_AsRoot(p), a.get_Container_ID(p), a.get_DiscardLog(p), a.get_ArtefactID(p), a.get_Created(p), a.get_InstancesMeansPerAutodeployer(p), p.ID)

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
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No ApplicationDefinition with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) ApplicationDefinition with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBApplicationDefinition) TryByID(ctx context.Context, p uint64) (*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) ApplicationDefinition with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBApplicationDefinition) ByIDs(ctx context.Context, p []uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBApplicationDefinition) All(ctx context.Context) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBApplicationDefinition" rows with matching DownloadURL
func (a *DBApplicationDefinition) ByDownloadURL(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDownloadURL"
	l, e := a.fromQuery(ctx, qn, "downloadurl = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDownloadURL: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching DownloadURL
func (a *DBApplicationDefinition) ByMultiDownloadURL(ctx context.Context, p []string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDownloadURL"
	l, e := a.fromQuery(ctx, qn, "downloadurl in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDownloadURL: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeDownloadURL(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeDownloadURL"
	l, e := a.fromQuery(ctx, qn, "downloadurl ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDownloadURL: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching DownloadUser
func (a *DBApplicationDefinition) ByDownloadUser(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDownloadUser"
	l, e := a.fromQuery(ctx, qn, "downloaduser = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDownloadUser: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching DownloadUser
func (a *DBApplicationDefinition) ByMultiDownloadUser(ctx context.Context, p []string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDownloadUser"
	l, e := a.fromQuery(ctx, qn, "downloaduser in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDownloadUser: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeDownloadUser(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeDownloadUser"
	l, e := a.fromQuery(ctx, qn, "downloaduser ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDownloadUser: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching DownloadPassword
func (a *DBApplicationDefinition) ByDownloadPassword(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDownloadPassword"
	l, e := a.fromQuery(ctx, qn, "downloadpassword = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDownloadPassword: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching DownloadPassword
func (a *DBApplicationDefinition) ByMultiDownloadPassword(ctx context.Context, p []string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDownloadPassword"
	l, e := a.fromQuery(ctx, qn, "downloadpassword in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDownloadPassword: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeDownloadPassword(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeDownloadPassword"
	l, e := a.fromQuery(ctx, qn, "downloadpassword ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDownloadPassword: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Binary
func (a *DBApplicationDefinition) ByBinary(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByBinary"
	l, e := a.fromQuery(ctx, qn, "r_binary = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBinary: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching Binary
func (a *DBApplicationDefinition) ByMultiBinary(ctx context.Context, p []string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByBinary"
	l, e := a.fromQuery(ctx, qn, "r_binary in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBinary: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeBinary(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeBinary"
	l, e := a.fromQuery(ctx, qn, "r_binary ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBinary: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching BuildID
func (a *DBApplicationDefinition) ByBuildID(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByBuildID"
	l, e := a.fromQuery(ctx, qn, "buildid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching BuildID
func (a *DBApplicationDefinition) ByMultiBuildID(ctx context.Context, p []uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByBuildID"
	l, e := a.fromQuery(ctx, qn, "buildid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeBuildID(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeBuildID"
	l, e := a.fromQuery(ctx, qn, "buildid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Instances
func (a *DBApplicationDefinition) ByInstances(ctx context.Context, p uint32) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByInstances"
	l, e := a.fromQuery(ctx, qn, "instances = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByInstances: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching Instances
func (a *DBApplicationDefinition) ByMultiInstances(ctx context.Context, p []uint32) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByInstances"
	l, e := a.fromQuery(ctx, qn, "instances in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByInstances: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeInstances(ctx context.Context, p uint32) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeInstances"
	l, e := a.fromQuery(ctx, qn, "instances ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByInstances: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching DeploymentID
func (a *DBApplicationDefinition) ByDeploymentID(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDeploymentID"
	l, e := a.fromQuery(ctx, qn, "deploymentid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeploymentID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching DeploymentID
func (a *DBApplicationDefinition) ByMultiDeploymentID(ctx context.Context, p []string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDeploymentID"
	l, e := a.fromQuery(ctx, qn, "deploymentid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeploymentID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeDeploymentID(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeDeploymentID"
	l, e := a.fromQuery(ctx, qn, "deploymentid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeploymentID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Machines
func (a *DBApplicationDefinition) ByMachines(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByMachines"
	l, e := a.fromQuery(ctx, qn, "machines = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMachines: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching Machines
func (a *DBApplicationDefinition) ByMultiMachines(ctx context.Context, p []string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByMachines"
	l, e := a.fromQuery(ctx, qn, "machines in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMachines: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeMachines(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeMachines"
	l, e := a.fromQuery(ctx, qn, "machines ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMachines: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching DeployType
func (a *DBApplicationDefinition) ByDeployType(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDeployType"
	l, e := a.fromQuery(ctx, qn, "deploytype = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeployType: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching DeployType
func (a *DBApplicationDefinition) ByMultiDeployType(ctx context.Context, p []string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDeployType"
	l, e := a.fromQuery(ctx, qn, "deploytype in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeployType: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeDeployType(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeDeployType"
	l, e := a.fromQuery(ctx, qn, "deploytype ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeployType: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Critical
func (a *DBApplicationDefinition) ByCritical(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByCritical"
	l, e := a.fromQuery(ctx, qn, "critical = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCritical: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching Critical
func (a *DBApplicationDefinition) ByMultiCritical(ctx context.Context, p []bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByCritical"
	l, e := a.fromQuery(ctx, qn, "critical in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCritical: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeCritical(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeCritical"
	l, e := a.fromQuery(ctx, qn, "critical ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCritical: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching AlwaysOn
func (a *DBApplicationDefinition) ByAlwaysOn(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByAlwaysOn"
	l, e := a.fromQuery(ctx, qn, "alwayson = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAlwaysOn: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching AlwaysOn
func (a *DBApplicationDefinition) ByMultiAlwaysOn(ctx context.Context, p []bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByAlwaysOn"
	l, e := a.fromQuery(ctx, qn, "alwayson in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAlwaysOn: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeAlwaysOn(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeAlwaysOn"
	l, e := a.fromQuery(ctx, qn, "alwayson ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAlwaysOn: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching StaticTargetDir
func (a *DBApplicationDefinition) ByStaticTargetDir(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByStaticTargetDir"
	l, e := a.fromQuery(ctx, qn, "statictargetdir = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByStaticTargetDir: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching StaticTargetDir
func (a *DBApplicationDefinition) ByMultiStaticTargetDir(ctx context.Context, p []string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByStaticTargetDir"
	l, e := a.fromQuery(ctx, qn, "statictargetdir in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByStaticTargetDir: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeStaticTargetDir(ctx context.Context, p string) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeStaticTargetDir"
	l, e := a.fromQuery(ctx, qn, "statictargetdir ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByStaticTargetDir: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Public
func (a *DBApplicationDefinition) ByPublic(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByPublic"
	l, e := a.fromQuery(ctx, qn, "r_public = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPublic: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching Public
func (a *DBApplicationDefinition) ByMultiPublic(ctx context.Context, p []bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByPublic"
	l, e := a.fromQuery(ctx, qn, "r_public in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPublic: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikePublic(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikePublic"
	l, e := a.fromQuery(ctx, qn, "r_public ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPublic: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Java
func (a *DBApplicationDefinition) ByJava(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByJava"
	l, e := a.fromQuery(ctx, qn, "java = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByJava: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching Java
func (a *DBApplicationDefinition) ByMultiJava(ctx context.Context, p []bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByJava"
	l, e := a.fromQuery(ctx, qn, "java in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByJava: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeJava(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeJava"
	l, e := a.fromQuery(ctx, qn, "java ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByJava: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching RepositoryID
func (a *DBApplicationDefinition) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching RepositoryID
func (a *DBApplicationDefinition) ByMultiRepositoryID(ctx context.Context, p []uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching AsRoot
func (a *DBApplicationDefinition) ByAsRoot(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByAsRoot"
	l, e := a.fromQuery(ctx, qn, "asroot = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAsRoot: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching AsRoot
func (a *DBApplicationDefinition) ByMultiAsRoot(ctx context.Context, p []bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByAsRoot"
	l, e := a.fromQuery(ctx, qn, "asroot in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAsRoot: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeAsRoot(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeAsRoot"
	l, e := a.fromQuery(ctx, qn, "asroot ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAsRoot: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Container
func (a *DBApplicationDefinition) ByContainer(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByContainer"
	l, e := a.fromQuery(ctx, qn, "container = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByContainer: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching Container
func (a *DBApplicationDefinition) ByMultiContainer(ctx context.Context, p []uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByContainer"
	l, e := a.fromQuery(ctx, qn, "container in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByContainer: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeContainer(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeContainer"
	l, e := a.fromQuery(ctx, qn, "container ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByContainer: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching DiscardLog
func (a *DBApplicationDefinition) ByDiscardLog(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDiscardLog"
	l, e := a.fromQuery(ctx, qn, "discardlog = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDiscardLog: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching DiscardLog
func (a *DBApplicationDefinition) ByMultiDiscardLog(ctx context.Context, p []bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByDiscardLog"
	l, e := a.fromQuery(ctx, qn, "discardlog in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDiscardLog: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeDiscardLog(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeDiscardLog"
	l, e := a.fromQuery(ctx, qn, "discardlog ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDiscardLog: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching ArtefactID
func (a *DBApplicationDefinition) ByArtefactID(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByArtefactID"
	l, e := a.fromQuery(ctx, qn, "artefactid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByArtefactID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching ArtefactID
func (a *DBApplicationDefinition) ByMultiArtefactID(ctx context.Context, p []uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByArtefactID"
	l, e := a.fromQuery(ctx, qn, "artefactid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByArtefactID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeArtefactID(ctx context.Context, p uint64) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeArtefactID"
	l, e := a.fromQuery(ctx, qn, "artefactid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByArtefactID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching Created
func (a *DBApplicationDefinition) ByCreated(ctx context.Context, p uint32) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByCreated"
	l, e := a.fromQuery(ctx, qn, "created = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching Created
func (a *DBApplicationDefinition) ByMultiCreated(ctx context.Context, p []uint32) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByCreated"
	l, e := a.fromQuery(ctx, qn, "created in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreated: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeCreated(ctx context.Context, p uint32) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeCreated"
	l, e := a.fromQuery(ctx, qn, "created ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with matching InstancesMeansPerAutodeployer
func (a *DBApplicationDefinition) ByInstancesMeansPerAutodeployer(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByInstancesMeansPerAutodeployer"
	l, e := a.fromQuery(ctx, qn, "instancesmeansperautodeployer = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByInstancesMeansPerAutodeployer: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplicationDefinition" rows with multiple matching InstancesMeansPerAutodeployer
func (a *DBApplicationDefinition) ByMultiInstancesMeansPerAutodeployer(ctx context.Context, p []bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByInstancesMeansPerAutodeployer"
	l, e := a.fromQuery(ctx, qn, "instancesmeansperautodeployer in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByInstancesMeansPerAutodeployer: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplicationDefinition) ByLikeInstancesMeansPerAutodeployer(ctx context.Context, p bool) ([]*savepb.ApplicationDefinition, error) {
	qn := "DBApplicationDefinition_ByLikeInstancesMeansPerAutodeployer"
	l, e := a.fromQuery(ctx, qn, "instancesmeansperautodeployer ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByInstancesMeansPerAutodeployer: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBApplicationDefinition) get_ID(p *savepb.ApplicationDefinition) uint64 {
	return uint64(p.ID)
}

// getter for field "DownloadURL" (DownloadURL) [string]
func (a *DBApplicationDefinition) get_DownloadURL(p *savepb.ApplicationDefinition) string {
	return string(p.DownloadURL)
}

// getter for field "DownloadUser" (DownloadUser) [string]
func (a *DBApplicationDefinition) get_DownloadUser(p *savepb.ApplicationDefinition) string {
	return string(p.DownloadUser)
}

// getter for field "DownloadPassword" (DownloadPassword) [string]
func (a *DBApplicationDefinition) get_DownloadPassword(p *savepb.ApplicationDefinition) string {
	return string(p.DownloadPassword)
}

// getter for field "Binary" (Binary) [string]
func (a *DBApplicationDefinition) get_Binary(p *savepb.ApplicationDefinition) string {
	return string(p.Binary)
}

// getter for field "BuildID" (BuildID) [uint64]
func (a *DBApplicationDefinition) get_BuildID(p *savepb.ApplicationDefinition) uint64 {
	return uint64(p.BuildID)
}

// getter for field "Instances" (Instances) [uint32]
func (a *DBApplicationDefinition) get_Instances(p *savepb.ApplicationDefinition) uint32 {
	return uint32(p.Instances)
}

// getter for field "DeploymentID" (DeploymentID) [string]
func (a *DBApplicationDefinition) get_DeploymentID(p *savepb.ApplicationDefinition) string {
	return string(p.DeploymentID)
}

// getter for field "Machines" (Machines) [string]
func (a *DBApplicationDefinition) get_Machines(p *savepb.ApplicationDefinition) string {
	return string(p.Machines)
}

// getter for field "DeployType" (DeployType) [string]
func (a *DBApplicationDefinition) get_DeployType(p *savepb.ApplicationDefinition) string {
	return string(p.DeployType)
}

// getter for field "Critical" (Critical) [bool]
func (a *DBApplicationDefinition) get_Critical(p *savepb.ApplicationDefinition) bool {
	return bool(p.Critical)
}

// getter for field "AlwaysOn" (AlwaysOn) [bool]
func (a *DBApplicationDefinition) get_AlwaysOn(p *savepb.ApplicationDefinition) bool {
	return bool(p.AlwaysOn)
}

// getter for field "StaticTargetDir" (StaticTargetDir) [string]
func (a *DBApplicationDefinition) get_StaticTargetDir(p *savepb.ApplicationDefinition) string {
	return string(p.StaticTargetDir)
}

// getter for field "Public" (Public) [bool]
func (a *DBApplicationDefinition) get_Public(p *savepb.ApplicationDefinition) bool {
	return bool(p.Public)
}

// getter for field "Java" (Java) [bool]
func (a *DBApplicationDefinition) get_Java(p *savepb.ApplicationDefinition) bool {
	return bool(p.Java)
}

// getter for field "RepositoryID" (RepositoryID) [uint64]
func (a *DBApplicationDefinition) get_RepositoryID(p *savepb.ApplicationDefinition) uint64 {
	return uint64(p.RepositoryID)
}

// getter for field "AsRoot" (AsRoot) [bool]
func (a *DBApplicationDefinition) get_AsRoot(p *savepb.ApplicationDefinition) bool {
	return bool(p.AsRoot)
}

// getter for reference "Container"
func (a *DBApplicationDefinition) get_Container_ID(p *savepb.ApplicationDefinition) gosql.NullInt64 {
	if p.Container == nil {
		return gosql.NullInt64{Valid: false}
	}
	return gosql.NullInt64{Valid: true, Int64: int64(p.Container.ID)}
}

// getter for field "DiscardLog" (DiscardLog) [bool]
func (a *DBApplicationDefinition) get_DiscardLog(p *savepb.ApplicationDefinition) bool {
	return bool(p.DiscardLog)
}

// getter for field "ArtefactID" (ArtefactID) [uint64]
func (a *DBApplicationDefinition) get_ArtefactID(p *savepb.ApplicationDefinition) uint64 {
	return uint64(p.ArtefactID)
}

// getter for field "Created" (Created) [uint32]
func (a *DBApplicationDefinition) get_Created(p *savepb.ApplicationDefinition) uint32 {
	return uint32(p.Created)
}

// getter for field "InstancesMeansPerAutodeployer" (InstancesMeansPerAutodeployer) [bool]
func (a *DBApplicationDefinition) get_InstancesMeansPerAutodeployer(p *savepb.ApplicationDefinition) bool {
	return bool(p.InstancesMeansPerAutodeployer)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBApplicationDefinition) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.ApplicationDefinition, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	i := 0
	for col_name, value := range extra_fields {
		i++
		efname := fmt.Sprintf("EXTRA_FIELD_%d", i)
		query.Add(col_name+" = "+efname, QP{efname: value})
	}

	gw, paras := query.ToPostgres()
	queryname := "custom_dbquery"
	rows, err := a.DB.QueryContext(ctx, queryname, "select "+a.SelectCols()+" from "+a.Tablename()+" where "+gw, paras...)
	if err != nil {
		return nil, err
	}
	res, err := a.FromRows(ctx, rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	return res, nil

}

func (a *DBApplicationDefinition) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.ApplicationDefinition, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBApplicationDefinition) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.ApplicationDefinition, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	eq := ""
	if extra_fields != nil && len(extra_fields) > 0 {
		eq = " AND ("
		// build the extraquery "eq"
		i := len(args)
		deli := ""
		for col_name, value := range extra_fields {
			i++
			eq = eq + deli + col_name + fmt.Sprintf(" = $%d", i)
			deli = " AND "
			args = append(args, value)
		}
		eq = eq + ")"
	}
	rows, err := a.DB.QueryContext(ctx, queryname, "select "+a.SelectCols()+" from "+a.Tablename()+" where ( "+query_where+") "+eq, args...)
	if err != nil {
		return nil, err
	}
	res, err := a.FromRows(ctx, rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	return res, nil
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBApplicationDefinition) get_col_from_proto(p *savepb.ApplicationDefinition, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "downloadurl" {
		return a.get_DownloadURL(p)
	} else if colname == "downloaduser" {
		return a.get_DownloadUser(p)
	} else if colname == "downloadpassword" {
		return a.get_DownloadPassword(p)
	} else if colname == "r_binary" {
		return a.get_Binary(p)
	} else if colname == "buildid" {
		return a.get_BuildID(p)
	} else if colname == "instances" {
		return a.get_Instances(p)
	} else if colname == "deploymentid" {
		return a.get_DeploymentID(p)
	} else if colname == "machines" {
		return a.get_Machines(p)
	} else if colname == "deploytype" {
		return a.get_DeployType(p)
	} else if colname == "critical" {
		return a.get_Critical(p)
	} else if colname == "alwayson" {
		return a.get_AlwaysOn(p)
	} else if colname == "statictargetdir" {
		return a.get_StaticTargetDir(p)
	} else if colname == "r_public" {
		return a.get_Public(p)
	} else if colname == "java" {
		return a.get_Java(p)
	} else if colname == "repositoryid" {
		return a.get_RepositoryID(p)
	} else if colname == "asroot" {
		return a.get_AsRoot(p)
	} else if colname == "container" {
		return a.get_Container_ID(p)
	} else if colname == "discardlog" {
		return a.get_DiscardLog(p)
	} else if colname == "artefactid" {
		return a.get_ArtefactID(p)
	} else if colname == "created" {
		return a.get_Created(p)
	} else if colname == "instancesmeansperautodeployer" {
		return a.get_InstancesMeansPerAutodeployer(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBApplicationDefinition) Tablename() string {
	return a.SQLTablename
}

func (a *DBApplicationDefinition) SelectCols() string {
	return "id,downloadurl, downloaduser, downloadpassword, r_binary, buildid, instances, deploymentid, machines, deploytype, critical, alwayson, statictargetdir, r_public, java, repositoryid, asroot, container, discardlog, artefactid, created, instancesmeansperautodeployer"
}
func (a *DBApplicationDefinition) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".downloadurl, " + a.SQLTablename + ".downloaduser, " + a.SQLTablename + ".downloadpassword, " + a.SQLTablename + ".r_binary, " + a.SQLTablename + ".buildid, " + a.SQLTablename + ".instances, " + a.SQLTablename + ".deploymentid, " + a.SQLTablename + ".machines, " + a.SQLTablename + ".deploytype, " + a.SQLTablename + ".critical, " + a.SQLTablename + ".alwayson, " + a.SQLTablename + ".statictargetdir, " + a.SQLTablename + ".r_public, " + a.SQLTablename + ".java, " + a.SQLTablename + ".repositoryid, " + a.SQLTablename + ".asroot, " + a.SQLTablename + ".container, " + a.SQLTablename + ".discardlog, " + a.SQLTablename + ".artefactid, " + a.SQLTablename + ".created, " + a.SQLTablename + ".instancesmeansperautodeployer"
}

func (a *DBApplicationDefinition) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.ApplicationDefinition, error) {
	var res []*savepb.ApplicationDefinition
	for rows.Next() {
		// SCANNER:
		foo := &savepb.ApplicationDefinition{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.DownloadURL
		scanTarget_2 := &foo.DownloadUser
		scanTarget_3 := &foo.DownloadPassword
		scanTarget_4 := &foo.Binary
		scanTarget_5 := &foo.BuildID
		scanTarget_6 := &foo.Instances
		scanTarget_7 := &foo.DeploymentID
		scanTarget_8 := &foo.Machines
		scanTarget_9 := &foo.DeployType
		scanTarget_10 := &foo.Critical
		scanTarget_11 := &foo.AlwaysOn
		scanTarget_12 := &foo.StaticTargetDir
		scanTarget_13 := &foo.Public
		scanTarget_14 := &foo.Java
		scanTarget_15 := &foo.RepositoryID
		scanTarget_16 := &foo.AsRoot
		scanTarget_17 := &gosql.NullInt64{} // Container.ID
		scanTarget_18 := &foo.DiscardLog
		scanTarget_19 := &foo.ArtefactID
		scanTarget_20 := &foo.Created
		scanTarget_21 := &foo.InstancesMeansPerAutodeployer
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4, scanTarget_5, scanTarget_6, scanTarget_7, scanTarget_8, scanTarget_9, scanTarget_10, scanTarget_11, scanTarget_12, scanTarget_13, scanTarget_14, scanTarget_15, scanTarget_16, scanTarget_17, scanTarget_18, scanTarget_19, scanTarget_20, scanTarget_21)
		if scanTarget_17.Valid {
			if foo.Container == nil {
				foo.Container = &savepb.ContainerDef{}
			}

			_, err := scanTarget_17.Value()
			if err != nil {
				return nil, err
			}
			foo.Container.ID = uint64(scanTarget_17.Int64)
		}
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
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),downloadurl text not null ,downloaduser text not null ,downloadpassword text not null ,r_binary text not null ,buildid bigint not null ,instances integer not null ,deploymentid text not null ,machines text not null ,deploytype text not null ,critical boolean not null ,alwayson boolean not null ,statictargetdir text not null ,r_public boolean not null ,java boolean not null ,repositoryid bigint not null ,asroot boolean not null ,container bigint  ,discardlog boolean not null ,artefactid bigint not null ,created integer not null ,instancesmeansperautodeployer boolean not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),downloadurl text not null ,downloaduser text not null ,downloadpassword text not null ,r_binary text not null ,buildid bigint not null ,instances integer not null ,deploymentid text not null ,machines text not null ,deploytype text not null ,critical boolean not null ,alwayson boolean not null ,statictargetdir text not null ,r_public boolean not null ,java boolean not null ,repositoryid bigint not null ,asroot boolean not null ,container bigint  ,discardlog boolean not null ,artefactid bigint not null ,created integer not null ,instancesmeansperautodeployer boolean not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS downloadurl text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS downloaduser text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS downloadpassword text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS r_binary text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS buildid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS instances integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS deploymentid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS machines text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS deploytype text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS critical boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS alwayson boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS statictargetdir text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS r_public boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS java boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS asroot boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS container bigint  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS discardlog boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS artefactid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS created integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS instancesmeansperautodeployer boolean not null default false;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS downloadurl text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS downloaduser text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS downloadpassword text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS r_binary text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS buildid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS instances integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS deploymentid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS machines text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS deploytype text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS critical boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS alwayson boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS statictargetdir text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS r_public boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS java boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS repositoryid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS asroot boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS container bigint   default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS discardlog boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS artefactid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS created integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS instancesmeansperautodeployer boolean not null  default false;`,
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
		`ALTER TABLE ` + a.SQLTablename + ` add constraint mkdb_fk_applicationdefinition_container_containerdefid FOREIGN KEY (container) references containerdef (id) on delete cascade ;`,
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
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

