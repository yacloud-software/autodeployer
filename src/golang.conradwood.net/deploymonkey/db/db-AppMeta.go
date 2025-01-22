package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBAppMeta
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence appmeta_seq;

Main Table:

 CREATE TABLE appmeta (id integer primary key default nextval('appmeta_seq'),userrequestedstop boolean not null  ,created integer not null  ,app_id bigint not null  references applicationdefinition (id) on delete cascade  );

Alter statements:
ALTER TABLE appmeta ADD COLUMN IF NOT EXISTS userrequestedstop boolean not null default false;
ALTER TABLE appmeta ADD COLUMN IF NOT EXISTS created integer not null default 0;
ALTER TABLE appmeta ADD COLUMN IF NOT EXISTS app_id bigint not null references applicationdefinition (id) on delete cascade  default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE appmeta_archive (id integer unique not null,userrequestedstop boolean not null,created integer not null,app_id bigint not null);
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
	default_def_DBAppMeta *DBAppMeta
)

type DBAppMeta struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func DefaultDBAppMeta() *DBAppMeta {
	if default_def_DBAppMeta != nil {
		return default_def_DBAppMeta
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBAppMeta(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBAppMeta = res
	return res
}
func NewDBAppMeta(db *sql.DB) *DBAppMeta {
	foo := DBAppMeta{DB: db}
	foo.SQLTablename = "appmeta"
	foo.SQLArchivetablename = "appmeta_archive"
	return &foo
}

func (a *DBAppMeta) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBAppMeta) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

// archive. It is NOT transactionally save.
func (a *DBAppMeta) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBAppMeta", "insert into "+a.SQLArchivetablename+" (id,userrequestedstop, created, app_id) values ($1,$2, $3, $4) ", p.ID, p.UserRequestedStop, p.Created, p.AppDef.ID)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBAppMeta) buildSaveMap(ctx context.Context, p *savepb.AppMeta) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["userrequestedstop"] = a.get_col_from_proto(p, "userrequestedstop")
	res["created"] = a.get_col_from_proto(p, "created")
	res["app_id"] = a.get_col_from_proto(p, "app_id")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBAppMeta) Save(ctx context.Context, p *savepb.AppMeta) (uint64, error) {
	qn := "save_DBAppMeta"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBAppMeta) SaveWithID(ctx context.Context, p *savepb.AppMeta) error {
	qn := "insert_DBAppMeta"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBAppMeta) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.AppMeta) (uint64, error) {
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

func (a *DBAppMeta) Update(ctx context.Context, p *savepb.AppMeta) error {
	qn := "DBAppMeta_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set userrequestedstop=$1, created=$2, app_id=$3 where id = $4", a.get_UserRequestedStop(p), a.get_Created(p), a.get_AppDef_ID(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBAppMeta) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBAppMeta_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBAppMeta) ByID(ctx context.Context, p uint64) (*savepb.AppMeta, error) {
	qn := "DBAppMeta_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No AppMeta with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) AppMeta with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBAppMeta) TryByID(ctx context.Context, p uint64) (*savepb.AppMeta, error) {
	qn := "DBAppMeta_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) AppMeta with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBAppMeta) ByIDs(ctx context.Context, p []uint64) ([]*savepb.AppMeta, error) {
	qn := "DBAppMeta_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBAppMeta) All(ctx context.Context) ([]*savepb.AppMeta, error) {
	qn := "DBAppMeta_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBAppMeta" rows with matching UserRequestedStop
func (a *DBAppMeta) ByUserRequestedStop(ctx context.Context, p bool) ([]*savepb.AppMeta, error) {
	qn := "DBAppMeta_ByUserRequestedStop"
	l, e := a.fromQuery(ctx, qn, "userrequestedstop = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserRequestedStop: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppMeta" rows with multiple matching UserRequestedStop
func (a *DBAppMeta) ByMultiUserRequestedStop(ctx context.Context, p []bool) ([]*savepb.AppMeta, error) {
	qn := "DBAppMeta_ByUserRequestedStop"
	l, e := a.fromQuery(ctx, qn, "userrequestedstop in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserRequestedStop: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBAppMeta) ByLikeUserRequestedStop(ctx context.Context, p bool) ([]*savepb.AppMeta, error) {
	qn := "DBAppMeta_ByLikeUserRequestedStop"
	l, e := a.fromQuery(ctx, qn, "userrequestedstop ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserRequestedStop: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppMeta" rows with matching Created
func (a *DBAppMeta) ByCreated(ctx context.Context, p uint32) ([]*savepb.AppMeta, error) {
	qn := "DBAppMeta_ByCreated"
	l, e := a.fromQuery(ctx, qn, "created = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppMeta" rows with multiple matching Created
func (a *DBAppMeta) ByMultiCreated(ctx context.Context, p []uint32) ([]*savepb.AppMeta, error) {
	qn := "DBAppMeta_ByCreated"
	l, e := a.fromQuery(ctx, qn, "created in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreated: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBAppMeta) ByLikeCreated(ctx context.Context, p uint32) ([]*savepb.AppMeta, error) {
	qn := "DBAppMeta_ByLikeCreated"
	l, e := a.fromQuery(ctx, qn, "created ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppMeta" rows with matching AppDef
func (a *DBAppMeta) ByAppDef(ctx context.Context, p uint64) ([]*savepb.AppMeta, error) {
	qn := "DBAppMeta_ByAppDef"
	l, e := a.fromQuery(ctx, qn, "app_id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAppDef: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppMeta" rows with multiple matching AppDef
func (a *DBAppMeta) ByMultiAppDef(ctx context.Context, p []uint64) ([]*savepb.AppMeta, error) {
	qn := "DBAppMeta_ByAppDef"
	l, e := a.fromQuery(ctx, qn, "app_id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAppDef: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBAppMeta) ByLikeAppDef(ctx context.Context, p uint64) ([]*savepb.AppMeta, error) {
	qn := "DBAppMeta_ByLikeAppDef"
	l, e := a.fromQuery(ctx, qn, "app_id ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAppDef: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBAppMeta) get_ID(p *savepb.AppMeta) uint64 {
	return uint64(p.ID)
}

// getter for field "UserRequestedStop" (UserRequestedStop) [bool]
func (a *DBAppMeta) get_UserRequestedStop(p *savepb.AppMeta) bool {
	return bool(p.UserRequestedStop)
}

// getter for field "Created" (Created) [uint32]
func (a *DBAppMeta) get_Created(p *savepb.AppMeta) uint32 {
	return uint32(p.Created)
}

// getter for reference "AppDef"
func (a *DBAppMeta) get_AppDef_ID(p *savepb.AppMeta) uint64 {
	if p.AppDef == nil {
		panic("field AppDef must not be nil")
	}
	return p.AppDef.ID
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBAppMeta) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.AppMeta, error) {
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

func (a *DBAppMeta) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.AppMeta, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBAppMeta) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.AppMeta, error) {
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
func (a *DBAppMeta) get_col_from_proto(p *savepb.AppMeta, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "userrequestedstop" {
		return a.get_UserRequestedStop(p)
	} else if colname == "created" {
		return a.get_Created(p)
	} else if colname == "app_id" {
		return a.get_AppDef_ID(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBAppMeta) Tablename() string {
	return a.SQLTablename
}

func (a *DBAppMeta) SelectCols() string {
	return "id,userrequestedstop, created, app_id"
}
func (a *DBAppMeta) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".userrequestedstop, " + a.SQLTablename + ".created, " + a.SQLTablename + ".app_id"
}

func (a *DBAppMeta) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.AppMeta, error) {
	var res []*savepb.AppMeta
	for rows.Next() {
		// SCANNER:
		foo := &savepb.AppMeta{}
		// create the non-nullable pointers
		foo.AppDef = &savepb.ApplicationDefinition{} // non-nullable
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.UserRequestedStop
		scanTarget_2 := &foo.Created
		scanTarget_3 := &foo.AppDef.ID
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3)
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
func (a *DBAppMeta) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userrequestedstop boolean not null ,created integer not null ,app_id bigint not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userrequestedstop boolean not null ,created integer not null ,app_id bigint not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS userrequestedstop boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS created integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS app_id bigint not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS userrequestedstop boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS created integer not null  default 0;`,
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
		`ALTER TABLE ` + a.SQLTablename + ` add constraint mkdb_fk_appmeta_app_id_applicationdefinitionid FOREIGN KEY (app_id) references applicationdefinition (id) on delete cascade ;`,
	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBAppMeta) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

