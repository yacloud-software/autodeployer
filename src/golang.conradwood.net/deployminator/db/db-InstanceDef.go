package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBInstanceDef
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence deployminator_instancedef_seq;

Main Table:

 CREATE TABLE deployminator_instancedef (id integer primary key default nextval('deployminator_instancedef_seq'),deploymentid bigint not null  references deployminator_deploymentdescriptor (id) on delete cascade  ,machinegroup text not null  ,instances integer not null  ,instancecountispermachine boolean not null  );

Alter statements:
ALTER TABLE deployminator_instancedef ADD COLUMN deploymentid bigint not null references deployminator_deploymentdescriptor (id) on delete cascade  default 0;
ALTER TABLE deployminator_instancedef ADD COLUMN machinegroup text not null default '';
ALTER TABLE deployminator_instancedef ADD COLUMN instances integer not null default 0;
ALTER TABLE deployminator_instancedef ADD COLUMN instancecountispermachine boolean not null default false;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE deployminator_instancedef_archive (id integer unique not null,deploymentid bigint not null,machinegroup text not null,instances integer not null,instancecountispermachine boolean not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/deployminator"
	"golang.conradwood.net/go-easyops/sql"
)

type DBInstanceDef struct {
	DB *sql.DB
}

func NewDBInstanceDef(db *sql.DB) *DBInstanceDef {
	foo := DBInstanceDef{DB: db}
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBInstanceDef) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "insert_DBInstanceDef", "insert into deployminator_instancedef_archive (id,deploymentid, machinegroup, instances, instancecountispermachine) values ($1,$2, $3, $4, $5) ", p.ID, p.DeploymentID.ID, p.MachineGroup, p.Instances, p.InstanceCountIsPerMachine)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBInstanceDef) Save(ctx context.Context, p *savepb.InstanceDef) (uint64, error) {
	qn := "DBInstanceDef_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into deployminator_instancedef (deploymentid, machinegroup, instances, instancecountispermachine) values ($1, $2, $3, $4) returning id", p.DeploymentID.ID, p.MachineGroup, p.Instances, p.InstanceCountIsPerMachine)
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
func (a *DBInstanceDef) SaveWithID(ctx context.Context, p *savepb.InstanceDef) error {
	qn := "insert_DBInstanceDef"
	_, e := a.DB.ExecContext(ctx, qn, "insert into deployminator_instancedef (id,deploymentid, machinegroup, instances, instancecountispermachine) values ($1,$2, $3, $4, $5) ", p.ID, p.DeploymentID.ID, p.MachineGroup, p.Instances, p.InstanceCountIsPerMachine)
	return a.Error(ctx, qn, e)
}

func (a *DBInstanceDef) Update(ctx context.Context, p *savepb.InstanceDef) error {
	qn := "DBInstanceDef_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update deployminator_instancedef set deploymentid=$1, machinegroup=$2, instances=$3, instancecountispermachine=$4 where id = $5", p.DeploymentID.ID, p.MachineGroup, p.Instances, p.InstanceCountIsPerMachine, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBInstanceDef) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBInstanceDef_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from deployminator_instancedef where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBInstanceDef) ByID(ctx context.Context, p uint64) (*savepb.InstanceDef, error) {
	qn := "DBInstanceDef_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from deployminator_instancedef where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No InstanceDef with id %d", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) InstanceDef with id %d", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBInstanceDef) All(ctx context.Context) ([]*savepb.InstanceDef, error) {
	qn := "DBInstanceDef_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from deployminator_instancedef order by id")
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

// get all "DBInstanceDef" rows with matching DeploymentID
func (a *DBInstanceDef) ByDeploymentID(ctx context.Context, p uint64) ([]*savepb.InstanceDef, error) {
	qn := "DBInstanceDef_ByDeploymentID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from deployminator_instancedef where deploymentid = $1", p)
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
func (a *DBInstanceDef) ByLikeDeploymentID(ctx context.Context, p uint64) ([]*savepb.InstanceDef, error) {
	qn := "DBInstanceDef_ByLikeDeploymentID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from deployminator_instancedef where deploymentid ilike $1", p)
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

// get all "DBInstanceDef" rows with matching MachineGroup
func (a *DBInstanceDef) ByMachineGroup(ctx context.Context, p string) ([]*savepb.InstanceDef, error) {
	qn := "DBInstanceDef_ByMachineGroup"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from deployminator_instancedef where machinegroup = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMachineGroup: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMachineGroup: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBInstanceDef) ByLikeMachineGroup(ctx context.Context, p string) ([]*savepb.InstanceDef, error) {
	qn := "DBInstanceDef_ByLikeMachineGroup"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from deployminator_instancedef where machinegroup ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMachineGroup: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMachineGroup: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBInstanceDef" rows with matching Instances
func (a *DBInstanceDef) ByInstances(ctx context.Context, p uint32) ([]*savepb.InstanceDef, error) {
	qn := "DBInstanceDef_ByInstances"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from deployminator_instancedef where instances = $1", p)
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
func (a *DBInstanceDef) ByLikeInstances(ctx context.Context, p uint32) ([]*savepb.InstanceDef, error) {
	qn := "DBInstanceDef_ByLikeInstances"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from deployminator_instancedef where instances ilike $1", p)
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

// get all "DBInstanceDef" rows with matching InstanceCountIsPerMachine
func (a *DBInstanceDef) ByInstanceCountIsPerMachine(ctx context.Context, p bool) ([]*savepb.InstanceDef, error) {
	qn := "DBInstanceDef_ByInstanceCountIsPerMachine"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from deployminator_instancedef where instancecountispermachine = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInstanceCountIsPerMachine: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInstanceCountIsPerMachine: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBInstanceDef) ByLikeInstanceCountIsPerMachine(ctx context.Context, p bool) ([]*savepb.InstanceDef, error) {
	qn := "DBInstanceDef_ByLikeInstanceCountIsPerMachine"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from deployminator_instancedef where instancecountispermachine ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInstanceCountIsPerMachine: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInstanceCountIsPerMachine: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBInstanceDef) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.InstanceDef, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBInstanceDef) Tablename() string {
	return "deployminator_instancedef"
}

func (a *DBInstanceDef) SelectCols() string {
	return "id,deploymentid, machinegroup, instances, instancecountispermachine"
}
func (a *DBInstanceDef) SelectColsQualified() string {
	return "deployminator_instancedef.id,deployminator_instancedef.deploymentid, deployminator_instancedef.machinegroup, deployminator_instancedef.instances, deployminator_instancedef.instancecountispermachine"
}

func (a *DBInstanceDef) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.InstanceDef, error) {
	var res []*savepb.InstanceDef
	for rows.Next() {
		foo := savepb.InstanceDef{DeploymentID: &savepb.DeploymentDescriptor{}}
		err := rows.Scan(&foo.ID, &foo.DeploymentID.ID, &foo.MachineGroup, &foo.Instances, &foo.InstanceCountIsPerMachine)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}

/**********************************************************************
* Helper to create table and columns
**********************************************************************/
func (a *DBInstanceDef) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence deployminator_instancedef_seq;`,
		`CREATE TABLE deployminator_instancedef (id integer primary key default nextval('deployminator_instancedef_seq'),deploymentid bigint not null,machinegroup text not null,instances integer not null,instancecountispermachine boolean not null);`,
	}
	for i, c := range csql {
		_, e := a.DB.ExecContext(ctx, fmt.Sprintf("create_deployminator_instancedef_%d", i), c)
		if e != nil {
			return e
		}
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBInstanceDef) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table=deployminator_instancedef, query=%s] Error: %s", q, e)
}
