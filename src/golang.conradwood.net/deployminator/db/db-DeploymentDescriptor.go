package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBDeploymentDescriptor
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence deployminator_deploymentdescriptor_seq;

Main Table:

 CREATE TABLE deployminator_deploymentdescriptor (id integer primary key default nextval('deployminator_deploymentdescriptor_seq'),application bigint not null  references deployminator_application (id) on delete cascade  ,buildnumber bigint not null  ,branch text not null  ,deployme boolean not null  );

Alter statements:
ALTER TABLE deployminator_deploymentdescriptor ADD COLUMN IF NOT EXISTS application bigint not null references deployminator_application (id) on delete cascade  default 0;
ALTER TABLE deployminator_deploymentdescriptor ADD COLUMN IF NOT EXISTS buildnumber bigint not null default 0;
ALTER TABLE deployminator_deploymentdescriptor ADD COLUMN IF NOT EXISTS branch text not null default '';
ALTER TABLE deployminator_deploymentdescriptor ADD COLUMN IF NOT EXISTS deployme boolean not null default false;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE deployminator_deploymentdescriptor_archive (id integer unique not null,application bigint not null,buildnumber bigint not null,branch text not null,deployme boolean not null);
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
	default_def_DBDeploymentDescriptor *DBDeploymentDescriptor
)

type DBDeploymentDescriptor struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBDeploymentDescriptor() *DBDeploymentDescriptor {
	if default_def_DBDeploymentDescriptor != nil {
		return default_def_DBDeploymentDescriptor
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBDeploymentDescriptor(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBDeploymentDescriptor = res
	return res
}
func NewDBDeploymentDescriptor(db *sql.DB) *DBDeploymentDescriptor {
	foo := DBDeploymentDescriptor{DB: db}
	foo.SQLTablename = "deployminator_deploymentdescriptor"
	foo.SQLArchivetablename = "deployminator_deploymentdescriptor_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBDeploymentDescriptor) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBDeploymentDescriptor", "insert into "+a.SQLArchivetablename+" (id,application, buildnumber, branch, deployme) values ($1,$2, $3, $4, $5) ", p.ID, p.Application.ID, p.BuildNumber, p.Branch, p.DeployMe)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBDeploymentDescriptor) Save(ctx context.Context, p *savepb.DeploymentDescriptor) (uint64, error) {
	qn := "DBDeploymentDescriptor_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (application, buildnumber, branch, deployme) values ($1, $2, $3, $4) returning id", p.Application.ID, p.BuildNumber, p.Branch, p.DeployMe)
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
func (a *DBDeploymentDescriptor) SaveWithID(ctx context.Context, p *savepb.DeploymentDescriptor) error {
	qn := "insert_DBDeploymentDescriptor"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,application, buildnumber, branch, deployme) values ($1,$2, $3, $4, $5) ", p.ID, p.Application.ID, p.BuildNumber, p.Branch, p.DeployMe)
	return a.Error(ctx, qn, e)
}

func (a *DBDeploymentDescriptor) Update(ctx context.Context, p *savepb.DeploymentDescriptor) error {
	qn := "DBDeploymentDescriptor_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set application=$1, buildnumber=$2, branch=$3, deployme=$4 where id = $5", p.Application.ID, p.BuildNumber, p.Branch, p.DeployMe, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBDeploymentDescriptor) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBDeploymentDescriptor_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBDeploymentDescriptor) ByID(ctx context.Context, p uint64) (*savepb.DeploymentDescriptor, error) {
	qn := "DBDeploymentDescriptor_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,application, buildnumber, branch, deployme from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No DeploymentDescriptor with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) DeploymentDescriptor with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBDeploymentDescriptor) TryByID(ctx context.Context, p uint64) (*savepb.DeploymentDescriptor, error) {
	qn := "DBDeploymentDescriptor_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,application, buildnumber, branch, deployme from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) DeploymentDescriptor with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBDeploymentDescriptor) All(ctx context.Context) ([]*savepb.DeploymentDescriptor, error) {
	qn := "DBDeploymentDescriptor_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,application, buildnumber, branch, deployme from "+a.SQLTablename+" order by id")
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

// get all "DBDeploymentDescriptor" rows with matching Application
func (a *DBDeploymentDescriptor) ByApplication(ctx context.Context, p uint64) ([]*savepb.DeploymentDescriptor, error) {
	qn := "DBDeploymentDescriptor_ByApplication"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,application, buildnumber, branch, deployme from "+a.SQLTablename+" where application = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByApplication: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByApplication: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentDescriptor) ByLikeApplication(ctx context.Context, p uint64) ([]*savepb.DeploymentDescriptor, error) {
	qn := "DBDeploymentDescriptor_ByLikeApplication"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,application, buildnumber, branch, deployme from "+a.SQLTablename+" where application ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByApplication: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByApplication: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentDescriptor" rows with matching BuildNumber
func (a *DBDeploymentDescriptor) ByBuildNumber(ctx context.Context, p uint64) ([]*savepb.DeploymentDescriptor, error) {
	qn := "DBDeploymentDescriptor_ByBuildNumber"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,application, buildnumber, branch, deployme from "+a.SQLTablename+" where buildnumber = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBuildNumber: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBuildNumber: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentDescriptor) ByLikeBuildNumber(ctx context.Context, p uint64) ([]*savepb.DeploymentDescriptor, error) {
	qn := "DBDeploymentDescriptor_ByLikeBuildNumber"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,application, buildnumber, branch, deployme from "+a.SQLTablename+" where buildnumber ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBuildNumber: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBuildNumber: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentDescriptor" rows with matching Branch
func (a *DBDeploymentDescriptor) ByBranch(ctx context.Context, p string) ([]*savepb.DeploymentDescriptor, error) {
	qn := "DBDeploymentDescriptor_ByBranch"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,application, buildnumber, branch, deployme from "+a.SQLTablename+" where branch = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBranch: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBranch: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentDescriptor) ByLikeBranch(ctx context.Context, p string) ([]*savepb.DeploymentDescriptor, error) {
	qn := "DBDeploymentDescriptor_ByLikeBranch"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,application, buildnumber, branch, deployme from "+a.SQLTablename+" where branch ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBranch: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBranch: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentDescriptor" rows with matching DeployMe
func (a *DBDeploymentDescriptor) ByDeployMe(ctx context.Context, p bool) ([]*savepb.DeploymentDescriptor, error) {
	qn := "DBDeploymentDescriptor_ByDeployMe"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,application, buildnumber, branch, deployme from "+a.SQLTablename+" where deployme = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployMe: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployMe: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentDescriptor) ByLikeDeployMe(ctx context.Context, p bool) ([]*savepb.DeploymentDescriptor, error) {
	qn := "DBDeploymentDescriptor_ByLikeDeployMe"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,application, buildnumber, branch, deployme from "+a.SQLTablename+" where deployme ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployMe: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployMe: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBDeploymentDescriptor) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.DeploymentDescriptor, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBDeploymentDescriptor) Tablename() string {
	return a.SQLTablename
}

func (a *DBDeploymentDescriptor) SelectCols() string {
	return "id,application, buildnumber, branch, deployme"
}
func (a *DBDeploymentDescriptor) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".application, " + a.SQLTablename + ".buildnumber, " + a.SQLTablename + ".branch, " + a.SQLTablename + ".deployme"
}

func (a *DBDeploymentDescriptor) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.DeploymentDescriptor, error) {
	var res []*savepb.DeploymentDescriptor
	for rows.Next() {
		foo := savepb.DeploymentDescriptor{Application: &savepb.Application{}}
		err := rows.Scan(&foo.ID, &foo.Application.ID, &foo.BuildNumber, &foo.Branch, &foo.DeployMe)
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
func (a *DBDeploymentDescriptor) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),application bigint not null  references deployminator_application (id) on delete cascade  ,buildnumber bigint not null  ,branch text not null  ,deployme boolean not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),application bigint not null  references deployminator_application (id) on delete cascade  ,buildnumber bigint not null  ,branch text not null  ,deployme boolean not null  );`,
		`ALTER TABLE deployminator_deploymentdescriptor ADD COLUMN IF NOT EXISTS application bigint not null references deployminator_application (id) on delete cascade  default 0;`,
		`ALTER TABLE deployminator_deploymentdescriptor ADD COLUMN IF NOT EXISTS buildnumber bigint not null default 0;`,
		`ALTER TABLE deployminator_deploymentdescriptor ADD COLUMN IF NOT EXISTS branch text not null default '';`,
		`ALTER TABLE deployminator_deploymentdescriptor ADD COLUMN IF NOT EXISTS deployme boolean not null default false;`,
	}
	for i, c := range csql {
		_, e := a.DB.ExecContext(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
		if e != nil {
			return e
		}
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBDeploymentDescriptor) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
