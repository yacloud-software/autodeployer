package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBContainerDef
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
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/sql"
	"os"
	"sync"
)

var (
	default_def_DBContainerDef *DBContainerDef
)

type DBContainerDef struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBContainerDef()
	})
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

func (a *DBContainerDef) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBContainerDef) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBContainerDef) NewQuery() *Query {
	return newQuery(a)
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

// return a map with columnname -> value_from_proto
func (a *DBContainerDef) buildSaveMap(ctx context.Context, p *savepb.ContainerDef) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["url"] = a.get_col_from_proto(p, "url")
	res["useoverlayfs"] = a.get_col_from_proto(p, "useoverlayfs")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBContainerDef) Save(ctx context.Context, p *savepb.ContainerDef) (uint64, error) {
	qn := "save_DBContainerDef"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBContainerDef) SaveWithID(ctx context.Context, p *savepb.ContainerDef) error {
	qn := "insert_DBContainerDef"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBContainerDef) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.ContainerDef) (uint64, error) {
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

// if ID==0 save, otherwise update
func (a *DBContainerDef) SaveOrUpdate(ctx context.Context, p *savepb.ContainerDef) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBContainerDef) Update(ctx context.Context, p *savepb.ContainerDef) error {
	qn := "DBContainerDef_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set url=$1, useoverlayfs=$2 where id = $3", a.get_URL(p), a.get_UseOverlayFS(p), p.ID)

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
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No ContainerDef with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) ContainerDef with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBContainerDef) TryByID(ctx context.Context, p uint64) (*savepb.ContainerDef, error) {
	qn := "DBContainerDef_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) ContainerDef with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBContainerDef) ByIDs(ctx context.Context, p []uint64) ([]*savepb.ContainerDef, error) {
	qn := "DBContainerDef_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBContainerDef) All(ctx context.Context) ([]*savepb.ContainerDef, error) {
	qn := "DBContainerDef_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBContainerDef" rows with matching URL
func (a *DBContainerDef) ByURL(ctx context.Context, p string) ([]*savepb.ContainerDef, error) {
	qn := "DBContainerDef_ByURL"
	l, e := a.fromQuery(ctx, qn, "url = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByURL: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBContainerDef" rows with multiple matching URL
func (a *DBContainerDef) ByMultiURL(ctx context.Context, p []string) ([]*savepb.ContainerDef, error) {
	qn := "DBContainerDef_ByURL"
	l, e := a.fromQuery(ctx, qn, "url in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByURL: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBContainerDef) ByLikeURL(ctx context.Context, p string) ([]*savepb.ContainerDef, error) {
	qn := "DBContainerDef_ByLikeURL"
	l, e := a.fromQuery(ctx, qn, "url ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByURL: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBContainerDef" rows with matching UseOverlayFS
func (a *DBContainerDef) ByUseOverlayFS(ctx context.Context, p bool) ([]*savepb.ContainerDef, error) {
	qn := "DBContainerDef_ByUseOverlayFS"
	l, e := a.fromQuery(ctx, qn, "useoverlayfs = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUseOverlayFS: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBContainerDef" rows with multiple matching UseOverlayFS
func (a *DBContainerDef) ByMultiUseOverlayFS(ctx context.Context, p []bool) ([]*savepb.ContainerDef, error) {
	qn := "DBContainerDef_ByUseOverlayFS"
	l, e := a.fromQuery(ctx, qn, "useoverlayfs in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUseOverlayFS: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBContainerDef) ByLikeUseOverlayFS(ctx context.Context, p bool) ([]*savepb.ContainerDef, error) {
	qn := "DBContainerDef_ByLikeUseOverlayFS"
	l, e := a.fromQuery(ctx, qn, "useoverlayfs ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUseOverlayFS: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBContainerDef) get_ID(p *savepb.ContainerDef) uint64 {
	return uint64(p.ID)
}

// getter for field "URL" (URL) [string]
func (a *DBContainerDef) get_URL(p *savepb.ContainerDef) string {
	return string(p.URL)
}

// getter for field "UseOverlayFS" (UseOverlayFS) [bool]
func (a *DBContainerDef) get_UseOverlayFS(p *savepb.ContainerDef) bool {
	return bool(p.UseOverlayFS)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBContainerDef) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.ContainerDef, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	i := 0
	for col_name, value := range extra_fields {
		i++
		/*
		   efname:=fmt.Sprintf("EXTRA_FIELD_%d",i)
		   query.Add(col_name+" = "+efname,QP{efname:value})
		*/
		query.AddEqual(col_name, value)
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

func (a *DBContainerDef) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.ContainerDef, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBContainerDef) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.ContainerDef, error) {
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
func (a *DBContainerDef) get_col_from_proto(p *savepb.ContainerDef, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "url" {
		return a.get_URL(p)
	} else if colname == "useoverlayfs" {
		return a.get_UseOverlayFS(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

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
		// SCANNER:
		foo := &savepb.ContainerDef{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.URL
		scanTarget_2 := &foo.UseOverlayFS
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
func (a *DBContainerDef) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),url text not null ,useoverlayfs boolean not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),url text not null ,useoverlayfs boolean not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS url text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS useoverlayfs boolean not null default false;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS url text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS useoverlayfs boolean not null  default false;`,
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
func (a *DBContainerDef) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

