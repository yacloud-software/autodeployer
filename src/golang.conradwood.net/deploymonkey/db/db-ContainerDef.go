package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBContainerDef
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence containerdef_seq;

Main Table:

 CREATE TABLE containerdef (id integer primary key default nextval('containerdef_seq'),url text not null  ,useoverlayfs boolean not null  );

Alter statements:
ALTER TABLE containerdef ADD COLUMN IF NOT EXISTS url text not null default '';
ALTER TABLE containerdef ADD COLUMN IF NOT EXISTS useoverlayfs boolean not null default false;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE containerdef_archive (id integer unique not null,url text not null,useoverlayfs boolean not null);
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
	default_def_DBContainerDef *DBContainerDef
)

type DBContainerDef struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBContainerDef() *DBContainerDef {
	if default_def_DBContainerDef != nil {
		return default_def_DBContainerDef
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBContainerDef(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBContainerDef = res
	return res
}
func NewDBContainerDef(db *sql.DB) *DBContainerDef {
	foo := DBContainerDef{DB: db}
	foo.SQLTablename = "containerdef"
	foo.SQLArchivetablename = "containerdef_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBContainerDef) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBContainerDef", "insert into "+a.SQLArchivetablename+" (id,url, useoverlayfs) values ($1,$2, $3) ", p.ID, p.URL, p.UseOverlayFS)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBContainerDef) Save(ctx context.Context, p *savepb.ContainerDef) (uint64, error) {
	qn := "DBContainerDef_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (url, useoverlayfs) values ($1, $2) returning id", p.URL, p.UseOverlayFS)
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
func (a *DBContainerDef) SaveWithID(ctx context.Context, p *savepb.ContainerDef) error {
	qn := "insert_DBContainerDef"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,url, useoverlayfs) values ($1,$2, $3) ", p.ID, p.URL, p.UseOverlayFS)
	return a.Error(ctx, qn, e)
}

func (a *DBContainerDef) Update(ctx context.Context, p *savepb.ContainerDef) error {
	qn := "DBContainerDef_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set url=$1, useoverlayfs=$2 where id = $3", p.URL, p.UseOverlayFS, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBContainerDef) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBContainerDef_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBContainerDef) ByID(ctx context.Context, p uint64) (*savepb.ContainerDef, error) {
	qn := "DBContainerDef_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,url, useoverlayfs from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No ContainerDef with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) ContainerDef with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBContainerDef) TryByID(ctx context.Context, p uint64) (*savepb.ContainerDef, error) {
	qn := "DBContainerDef_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,url, useoverlayfs from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) ContainerDef with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBContainerDef) All(ctx context.Context) ([]*savepb.ContainerDef, error) {
	qn := "DBContainerDef_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,url, useoverlayfs from "+a.SQLTablename+" order by id")
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

// get all "DBContainerDef" rows with matching URL
func (a *DBContainerDef) ByURL(ctx context.Context, p string) ([]*savepb.ContainerDef, error) {
	qn := "DBContainerDef_ByURL"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,url, useoverlayfs from "+a.SQLTablename+" where url = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByURL: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByURL: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBContainerDef) ByLikeURL(ctx context.Context, p string) ([]*savepb.ContainerDef, error) {
	qn := "DBContainerDef_ByLikeURL"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,url, useoverlayfs from "+a.SQLTablename+" where url ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByURL: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByURL: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBContainerDef" rows with matching UseOverlayFS
func (a *DBContainerDef) ByUseOverlayFS(ctx context.Context, p bool) ([]*savepb.ContainerDef, error) {
	qn := "DBContainerDef_ByUseOverlayFS"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,url, useoverlayfs from "+a.SQLTablename+" where useoverlayfs = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUseOverlayFS: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUseOverlayFS: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBContainerDef) ByLikeUseOverlayFS(ctx context.Context, p bool) ([]*savepb.ContainerDef, error) {
	qn := "DBContainerDef_ByLikeUseOverlayFS"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,url, useoverlayfs from "+a.SQLTablename+" where useoverlayfs ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUseOverlayFS: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUseOverlayFS: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBContainerDef) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.ContainerDef, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBContainerDef) Tablename() string {
	return a.SQLTablename
}

func (a *DBContainerDef) SelectCols() string {
	return "id,url, useoverlayfs"
}
func (a *DBContainerDef) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".url, " + a.SQLTablename + ".useoverlayfs"
}

func (a *DBContainerDef) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.ContainerDef, error) {
	var res []*savepb.ContainerDef
	for rows.Next() {
		foo := savepb.ContainerDef{}
		err := rows.Scan(&foo.ID, &foo.URL, &foo.UseOverlayFS)
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
func (a *DBContainerDef) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),url text not null  ,useoverlayfs boolean not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),url text not null  ,useoverlayfs boolean not null  );`,
		`ALTER TABLE containerdef ADD COLUMN IF NOT EXISTS url text not null default '';`,
		`ALTER TABLE containerdef ADD COLUMN IF NOT EXISTS useoverlayfs boolean not null default false;`,
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
func (a *DBContainerDef) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
