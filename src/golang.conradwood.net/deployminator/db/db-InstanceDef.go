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
ALTER TABLE deployminator_instancedef ADD COLUMN IF NOT EXISTS deploymentid bigint not null references deployminator_deploymentdescriptor (id) on delete cascade  default 0;
ALTER TABLE deployminator_instancedef ADD COLUMN IF NOT EXISTS machinegroup text not null default '';
ALTER TABLE deployminator_instancedef ADD COLUMN IF NOT EXISTS instances integer not null default 0;
ALTER TABLE deployminator_instancedef ADD COLUMN IF NOT EXISTS instancecountispermachine boolean not null default false;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE deployminator_instancedef_archive (id integer unique not null,deploymentid bigint not null,machinegroup text not null,instances integer not null,instancecountispermachine boolean not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/deployminator"
	"golang.conradwood.net/go-easyops/sql"
	"os"
)

var (
	default_def_DBInstanceDef *DBInstanceDef
)

type DBInstanceDef struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBInstanceDef() *DBInstanceDef {
	if default_def_DBInstanceDef != nil {
		return default_def_DBInstanceDef
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBInstanceDef(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBInstanceDef = res
	return res
}
func NewDBInstanceDef(db *sql.DB) *DBInstanceDef {
	foo := DBInstanceDef{DB: db}
	foo.SQLTablename = "deployminator_instancedef"
	foo.SQLArchivetablename = "deployminator_instancedef_archive"
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
	_, e := a.DB.ExecContext(ctx, "archive_DBInstanceDef", "insert into "+a.SQLArchivetablename+" (id,deploymentid, machinegroup, instances, instancecountispermachine) values ($1,$2, $3, $4, $5) ", p.ID, p.DeploymentID.ID, p.MachineGroup, p.Instances, p.InstanceCountIsPerMachine)
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
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (deploymentid, machinegroup, instances, instancecountispermachine) values ($1, $2, $3, $4) returning id", p.DeploymentID.ID, p.MachineGroup, p.Instances, p.InstanceCountIsPerMachine)
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
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,deploymentid, machinegroup, instances, instancecountispermachine) values ($1,$2, $3, $4, $5) ", p.ID, p.DeploymentID.ID, p.MachineGroup, p.Instances, p.InstanceCountIsPerMachine)
	return a.Error(ctx, qn, e)
}

func (a *DBInstanceDef) Update(ctx context.Context, p *savepb.InstanceDef) error {
	qn := "DBInstanceDef_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set deploymentid=$1, machinegroup=$2, instances=$3, instancecountispermachine=$4 where id = $5", p.DeploymentID.ID, p.MachineGroup, p.Instances, p.InstanceCountIsPerMachine, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBInstanceDef) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBInstanceDef_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBInstanceDef) ByID(ctx context.Context, p uint64) (*savepb.InstanceDef, error) {
	qn := "DBInstanceDef_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No InstanceDef with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) InstanceDef with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBInstanceDef) TryByID(ctx context.Context, p uint64) (*savepb.InstanceDef, error) {
	qn := "DBInstanceDef_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) InstanceDef with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBInstanceDef) All(ctx context.Context) ([]*savepb.InstanceDef, error) {
	qn := "DBInstanceDef_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from "+a.SQLTablename+" order by id")
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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from "+a.SQLTablename+" where deploymentid = $1", p)
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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from "+a.SQLTablename+" where deploymentid ilike $1", p)
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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from "+a.SQLTablename+" where machinegroup = $1", p)
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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from "+a.SQLTablename+" where machinegroup ilike $1", p)
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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from "+a.SQLTablename+" where instances = $1", p)
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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from "+a.SQLTablename+" where instances ilike $1", p)
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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from "+a.SQLTablename+" where instancecountispermachine = $1", p)
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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,deploymentid, machinegroup, instances, instancecountispermachine from "+a.SQLTablename+" where instancecountispermachine ilike $1", p)
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
	return a.SQLTablename
}

func (a *DBInstanceDef) SelectCols() string {
	return "id,deploymentid, machinegroup, instances, instancecountispermachine"
}
func (a *DBInstanceDef) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".deploymentid, " + a.SQLTablename + ".machinegroup, " + a.SQLTablename + ".instances, " + a.SQLTablename + ".instancecountispermachine"
}

func (a *DBInstanceDef) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.InstanceDef, error) {
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
func (a *DBInstanceDef) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.InstanceDef, error) {
	var res []*savepb.InstanceDef
	for rows.Next() {
		// SCANNER:
		foo := &savepb.InstanceDef{}
		// create the non-nullable pointers
		foo.DeploymentID = &savepb.DeploymentDescriptor{} // non-nullable
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.DeploymentID.ID
		scanTarget_2 := &foo.MachineGroup
		scanTarget_3 := &foo.Instances
		scanTarget_4 := &foo.InstanceCountIsPerMachine
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4)
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
func (a *DBInstanceDef) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),deploymentid bigint not null ,machinegroup text not null ,instances integer not null ,instancecountispermachine boolean not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),deploymentid bigint not null ,machinegroup text not null ,instances integer not null ,instancecountispermachine boolean not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS deploymentid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS machinegroup text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS instances integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS instancecountispermachine boolean not null default false;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS deploymentid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS machinegroup text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS instances integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS instancecountispermachine boolean not null  default false;`,
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
		`ALTER TABLE deployminator_instancedef add constraint mkdb_fk_4c2654b196d701919fb865962ded99a6 FOREIGN KEY (deploymentid) references deployminator_deploymentdescriptor (id) on delete cascade ;`,
	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
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
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

