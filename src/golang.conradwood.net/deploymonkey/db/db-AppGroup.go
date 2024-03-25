package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBAppGroup
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence appgroup_seq;

Main Table:

 CREATE TABLE appgroup (id integer primary key default nextval('appgroup_seq'),namespace text not null  ,groupname text not null  ,deployedversion integer not null  ,pendingversion integer not null  );

Alter statements:
ALTER TABLE appgroup ADD COLUMN IF NOT EXISTS namespace text not null default '';
ALTER TABLE appgroup ADD COLUMN IF NOT EXISTS groupname text not null default '';
ALTER TABLE appgroup ADD COLUMN IF NOT EXISTS deployedversion integer not null default 0;
ALTER TABLE appgroup ADD COLUMN IF NOT EXISTS pendingversion integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE appgroup_archive (id integer unique not null,namespace text not null,groupname text not null,deployedversion integer not null,pendingversion integer not null);
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
	default_def_DBAppGroup *DBAppGroup
)

type DBAppGroup struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBAppGroup() *DBAppGroup {
	if default_def_DBAppGroup != nil {
		return default_def_DBAppGroup
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBAppGroup(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBAppGroup = res
	return res
}
func NewDBAppGroup(db *sql.DB) *DBAppGroup {
	foo := DBAppGroup{DB: db}
	foo.SQLTablename = "appgroup"
	foo.SQLArchivetablename = "appgroup_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBAppGroup) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBAppGroup", "insert into "+a.SQLArchivetablename+" (id,namespace, groupname, deployedversion, pendingversion) values ($1,$2, $3, $4, $5) ", p.ID, p.Namespace, p.Groupname, p.DeployedVersion, p.PendingVersion)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBAppGroup) Save(ctx context.Context, p *savepb.AppGroup) (uint64, error) {
	qn := "DBAppGroup_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (namespace, groupname, deployedversion, pendingversion) values ($1, $2, $3, $4) returning id", a.get_Namespace(p), a.get_Groupname(p), a.get_DeployedVersion(p), a.get_PendingVersion(p))
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
func (a *DBAppGroup) SaveWithID(ctx context.Context, p *savepb.AppGroup) error {
	qn := "insert_DBAppGroup"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,namespace, groupname, deployedversion, pendingversion) values ($1,$2, $3, $4, $5) ", p.ID, p.Namespace, p.Groupname, p.DeployedVersion, p.PendingVersion)
	return a.Error(ctx, qn, e)
}

func (a *DBAppGroup) Update(ctx context.Context, p *savepb.AppGroup) error {
	qn := "DBAppGroup_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set namespace=$1, groupname=$2, deployedversion=$3, pendingversion=$4 where id = $5", a.get_Namespace(p), a.get_Groupname(p), a.get_DeployedVersion(p), a.get_PendingVersion(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBAppGroup) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBAppGroup_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBAppGroup) ByID(ctx context.Context, p uint64) (*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,namespace, groupname, deployedversion, pendingversion from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No AppGroup with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) AppGroup with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBAppGroup) TryByID(ctx context.Context, p uint64) (*savepb.AppGroup, error) {
	qn := "DBAppGroup_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,namespace, groupname, deployedversion, pendingversion from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) AppGroup with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBAppGroup) All(ctx context.Context) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,namespace, groupname, deployedversion, pendingversion from "+a.SQLTablename+" order by id")
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

// get all "DBAppGroup" rows with matching Namespace
func (a *DBAppGroup) ByNamespace(ctx context.Context, p string) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByNamespace"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,namespace, groupname, deployedversion, pendingversion from "+a.SQLTablename+" where namespace = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNamespace: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNamespace: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBAppGroup) ByLikeNamespace(ctx context.Context, p string) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByLikeNamespace"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,namespace, groupname, deployedversion, pendingversion from "+a.SQLTablename+" where namespace ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNamespace: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNamespace: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppGroup" rows with matching Groupname
func (a *DBAppGroup) ByGroupname(ctx context.Context, p string) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByGroupname"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,namespace, groupname, deployedversion, pendingversion from "+a.SQLTablename+" where groupname = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupname: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupname: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBAppGroup) ByLikeGroupname(ctx context.Context, p string) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByLikeGroupname"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,namespace, groupname, deployedversion, pendingversion from "+a.SQLTablename+" where groupname ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupname: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupname: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppGroup" rows with matching DeployedVersion
func (a *DBAppGroup) ByDeployedVersion(ctx context.Context, p uint32) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByDeployedVersion"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,namespace, groupname, deployedversion, pendingversion from "+a.SQLTablename+" where deployedversion = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployedVersion: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployedVersion: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBAppGroup) ByLikeDeployedVersion(ctx context.Context, p uint32) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByLikeDeployedVersion"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,namespace, groupname, deployedversion, pendingversion from "+a.SQLTablename+" where deployedversion ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployedVersion: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployedVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppGroup" rows with matching PendingVersion
func (a *DBAppGroup) ByPendingVersion(ctx context.Context, p uint32) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByPendingVersion"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,namespace, groupname, deployedversion, pendingversion from "+a.SQLTablename+" where pendingversion = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPendingVersion: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPendingVersion: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBAppGroup) ByLikePendingVersion(ctx context.Context, p uint32) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByLikePendingVersion"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,namespace, groupname, deployedversion, pendingversion from "+a.SQLTablename+" where pendingversion ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPendingVersion: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPendingVersion: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

func (a *DBAppGroup) get_ID(p *savepb.AppGroup) uint64 {
	return p.ID
}

func (a *DBAppGroup) get_Namespace(p *savepb.AppGroup) string {
	return p.Namespace
}

func (a *DBAppGroup) get_Groupname(p *savepb.AppGroup) string {
	return p.Groupname
}

func (a *DBAppGroup) get_DeployedVersion(p *savepb.AppGroup) uint32 {
	return p.DeployedVersion
}

func (a *DBAppGroup) get_PendingVersion(p *savepb.AppGroup) uint32 {
	return p.PendingVersion
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBAppGroup) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.AppGroup, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBAppGroup) Tablename() string {
	return a.SQLTablename
}

func (a *DBAppGroup) SelectCols() string {
	return "id,namespace, groupname, deployedversion, pendingversion"
}
func (a *DBAppGroup) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".namespace, " + a.SQLTablename + ".groupname, " + a.SQLTablename + ".deployedversion, " + a.SQLTablename + ".pendingversion"
}

func (a *DBAppGroup) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.AppGroup, error) {
	var res []*savepb.AppGroup
	for rows.Next() {
		foo := savepb.AppGroup{}
		err := rows.Scan(&foo.ID, &foo.Namespace, &foo.Groupname, &foo.DeployedVersion, &foo.PendingVersion)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}
func (a *DBAppGroup) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.AppGroup, error) {
	var res []*savepb.AppGroup
	for rows.Next() {
		// SCANNER:
		foo := &savepb.AppGroup{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.Namespace
		scanTarget_2 := &foo.Groupname
		scanTarget_3 := &foo.DeployedVersion
		scanTarget_4 := &foo.PendingVersion
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
func (a *DBAppGroup) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),namespace text not null ,groupname text not null ,deployedversion integer not null ,pendingversion integer not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),namespace text not null ,groupname text not null ,deployedversion integer not null ,pendingversion integer not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS namespace text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS groupname text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS deployedversion integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS pendingversion integer not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS namespace text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS groupname text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS deployedversion integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS pendingversion integer not null  default 0;`,
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

	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBAppGroup) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

