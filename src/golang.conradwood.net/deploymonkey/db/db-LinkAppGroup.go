package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBLinkAppGroup
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
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/sql"
	"os"
	"sync"
)

var (
	default_def_DBLinkAppGroup *DBLinkAppGroup
)

type DBLinkAppGroup struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBLinkAppGroup()
	})
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

func (a *DBLinkAppGroup) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBLinkAppGroup) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBLinkAppGroup) NewQuery() *Query {
	return newQuery(a)
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

// return a map with columnname -> value_from_proto
func (a *DBLinkAppGroup) buildSaveMap(ctx context.Context, p *savepb.LinkAppGroup) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["group_version_id"] = a.get_col_from_proto(p, "group_version_id")
	res["app_id"] = a.get_col_from_proto(p, "app_id")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBLinkAppGroup) Save(ctx context.Context, p *savepb.LinkAppGroup) (uint64, error) {
	qn := "save_DBLinkAppGroup"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBLinkAppGroup) SaveWithID(ctx context.Context, p *savepb.LinkAppGroup) error {
	qn := "insert_DBLinkAppGroup"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBLinkAppGroup) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.LinkAppGroup) (uint64, error) {
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
func (a *DBLinkAppGroup) SaveOrUpdate(ctx context.Context, p *savepb.LinkAppGroup) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
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
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No LinkAppGroup with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) LinkAppGroup with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBLinkAppGroup) TryByID(ctx context.Context, p uint64) (*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) LinkAppGroup with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBLinkAppGroup) ByIDs(ctx context.Context, p []uint64) ([]*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBLinkAppGroup) All(ctx context.Context) ([]*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBLinkAppGroup" rows with matching GroupVersion
func (a *DBLinkAppGroup) ByGroupVersion(ctx context.Context, p uint64) ([]*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_ByGroupVersion"
	l, e := a.fromQuery(ctx, qn, "group_version_id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByGroupVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBLinkAppGroup" rows with multiple matching GroupVersion
func (a *DBLinkAppGroup) ByMultiGroupVersion(ctx context.Context, p []uint64) ([]*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_ByGroupVersion"
	l, e := a.fromQuery(ctx, qn, "group_version_id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByGroupVersion: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBLinkAppGroup) ByLikeGroupVersion(ctx context.Context, p uint64) ([]*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_ByLikeGroupVersion"
	l, e := a.fromQuery(ctx, qn, "group_version_id ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByGroupVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBLinkAppGroup" rows with matching App
func (a *DBLinkAppGroup) ByApp(ctx context.Context, p uint64) ([]*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_ByApp"
	l, e := a.fromQuery(ctx, qn, "app_id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByApp: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBLinkAppGroup" rows with multiple matching App
func (a *DBLinkAppGroup) ByMultiApp(ctx context.Context, p []uint64) ([]*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_ByApp"
	l, e := a.fromQuery(ctx, qn, "app_id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByApp: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBLinkAppGroup) ByLikeApp(ctx context.Context, p uint64) ([]*savepb.LinkAppGroup, error) {
	qn := "DBLinkAppGroup_ByLikeApp"
	l, e := a.fromQuery(ctx, qn, "app_id ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByApp: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBLinkAppGroup) get_ID(p *savepb.LinkAppGroup) uint64 {
	return uint64(p.ID)
}

// getter for reference "GroupVersion"
func (a *DBLinkAppGroup) get_GroupVersion_ID(p *savepb.LinkAppGroup) uint64 {
	if p.GroupVersion == nil {
		panic("field GroupVersion must not be nil")
	}
	return p.GroupVersion.ID
}

// getter for reference "App"
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
func (a *DBLinkAppGroup) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.LinkAppGroup, error) {
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

func (a *DBLinkAppGroup) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.LinkAppGroup, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBLinkAppGroup) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.LinkAppGroup, error) {
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
func (a *DBLinkAppGroup) get_col_from_proto(p *savepb.LinkAppGroup, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "group_version_id" {
		return a.get_GroupVersion_ID(p)
	} else if colname == "app_id" {
		return a.get_App_ID(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBLinkAppGroup) Tablename() string {
	return a.SQLTablename
}

func (a *DBLinkAppGroup) SelectCols() string {
	return "id,group_version_id, app_id"
}
func (a *DBLinkAppGroup) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".group_version_id, " + a.SQLTablename + ".app_id"
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
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

