package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBLinkAppGroup
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence linkappgroup_seq;

Main Table:

 CREATE TABLE linkappgroup (id integer primary key default nextval('linkappgroup_seq'),group_version_id bigint not null  references group_version (id) on delete cascade  ,app_id bigint not null  references applicationdefinition (id) on delete cascade  );

Alter statements:
ALTER TABLE linkappgroup ADD COLUMN IF NOT EXISTS group_version_id bigint not null references group_version (id) on delete cascade  default 0;
ALTER TABLE linkappgroup ADD COLUMN IF NOT EXISTS app_id bigint not null references applicationdefinition (id) on delete cascade  default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE linkappgroup_archive (id integer unique not null,group_version_id bigint not null,app_id bigint not null);
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
	default_def_DBLinkAppGroup *DBLinkAppGroup
)

type DBLinkAppGroup struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBLinkAppGroup() *DBLinkAppGroup {
	if default_def_DBLinkAppGroup != nil {
		return default_def_DBLinkAppGroup
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBLinkAppGroup(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBLinkAppGroup = res
	return res
}
func NewDBLinkAppGroup(db *sql.DB) *DBLinkAppGroup {
	foo := DBLinkAppGroup{DB: db}
	foo.SQLTablename = "linkappgroup"
	foo.SQLArchivetablename = "linkappgroup_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBLinkAppGroup) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBLinkAppGroup", "insert into "+a.SQLArchivetablename+" (id,group_version_id, app_id) values ($1,$2, $3) ", p.ID, p.GroupVersion.ID, p.App.ID)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBLinkAppGroup) Save(ctx context.Context, p *savepb.LinkAppGroup) (uint64, error) {
	qn := "DBLinkAppGroup_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (group_version_id, app_id) values ($1, $2) returning id", a.get_GroupVersion_ID(p), a.get_App_ID(p))
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
func (a *DBLinkAppGroup) SaveWithID(ctx context.Context, p *savepb.LinkAppGroup) error {
	qn := "insert_DBLinkAppGroup"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,group_version_id, app_id) values ($1,$2, $3) ", p.ID, p.GroupVersion.ID, p.App.ID)
	return a.Error(ctx, qn, e)
}

func (a *DBLinkAppGroup) Update(ctx context.Context, p *savepb.LinkAppGroup) error {
	qn := "DBLinkAppGroup_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set group_version_id=$1, app_id=$2 where id = $3", a.get_GroupVersion_ID(p), a.get_App_ID(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBLinkAppGroup) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBLinkAppGroup_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBLinkAppGroup) ByID(ctx context.Context, p uint64) (*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,group_version_id, app_id from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No LinkAppGroup with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) LinkAppGroup with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBLinkAppGroup) TryByID(ctx context.Context, p uint64) (*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,group_version_id, app_id from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) LinkAppGroup with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBLinkAppGroup) All(ctx context.Context) ([]*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,group_version_id, app_id from "+a.SQLTablename+" order by id")
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

// get all "DBLinkAppGroup" rows with matching GroupVersion
func (a *DBLinkAppGroup) ByGroupVersion(ctx context.Context, p uint64) ([]*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_ByGroupVersion"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,group_version_id, app_id from "+a.SQLTablename+" where group_version_id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupVersion: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupVersion: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBLinkAppGroup) ByLikeGroupVersion(ctx context.Context, p uint64) ([]*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_ByLikeGroupVersion"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,group_version_id, app_id from "+a.SQLTablename+" where group_version_id ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupVersion: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBLinkAppGroup" rows with matching App
func (a *DBLinkAppGroup) ByApp(ctx context.Context, p uint64) ([]*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_ByApp"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,group_version_id, app_id from "+a.SQLTablename+" where app_id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByApp: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByApp: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBLinkAppGroup) ByLikeApp(ctx context.Context, p uint64) ([]*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_ByLikeApp"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,group_version_id, app_id from "+a.SQLTablename+" where app_id ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByApp: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByApp: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

func (a *DBLinkAppGroup) get_ID(p *savepb.LinkAppGroup) uint64 {
	return p.ID
}

func (a *DBLinkAppGroup) get_GroupVersion_ID(p *savepb.LinkAppGroup) uint64 {
	if p.GroupVersion == nil {
		panic("field GroupVersion must not be nil")
	}
	return p.GroupVersion.ID
}

func (a *DBLinkAppGroup) get_App_ID(p *savepb.LinkAppGroup) uint64 {
	if p.App == nil {
		panic("field App must not be nil")
	}
	return p.App.ID
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBLinkAppGroup) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.LinkAppGroup, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBLinkAppGroup) Tablename() string {
	return a.SQLTablename
}

func (a *DBLinkAppGroup) SelectCols() string {
	return "id,group_version_id, app_id"
}
func (a *DBLinkAppGroup) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".group_version_id, " + a.SQLTablename + ".app_id"
}

func (a *DBLinkAppGroup) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.LinkAppGroup, error) {
	var res []*savepb.LinkAppGroup
	for rows.Next() {
		foo := savepb.LinkAppGroup{GroupVersion: &savepb.GroupVersion{}, App: &savepb.ApplicationDefinition{}}
		err := rows.Scan(&foo.ID, &foo.GroupVersion.ID, &foo.App.ID)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}
func (a *DBLinkAppGroup) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.LinkAppGroup, error) {
	var res []*savepb.LinkAppGroup
	for rows.Next() {
		// SCANNER:
		foo := &savepb.LinkAppGroup{}
		// create the non-nullable pointers
		foo.GroupVersion = &savepb.GroupVersion{} // non-nullable
		foo.App = &savepb.ApplicationDefinition{} // non-nullable
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.GroupVersion.ID
		scanTarget_2 := &foo.App.ID
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2)
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
func (a *DBLinkAppGroup) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),group_version_id bigint not null ,app_id bigint not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),group_version_id bigint not null ,app_id bigint not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS group_version_id bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS app_id bigint not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS group_version_id bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS app_id bigint not null  default 0;`,
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
		`ALTER TABLE ` + a.SQLTablename + ` add constraint mkdb_fk_linkappgroup_group_version_id_group_versionid FOREIGN KEY (group_version_id) references group_version (id) on delete cascade ;`,
		`ALTER TABLE ` + a.SQLTablename + ` add constraint mkdb_fk_linkappgroup_app_id_applicationdefinitionid FOREIGN KEY (app_id) references applicationdefinition (id) on delete cascade ;`,
	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBLinkAppGroup) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

