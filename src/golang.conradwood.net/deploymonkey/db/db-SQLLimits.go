package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBSQLLimits
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence applimits_seq;

Main Table:

 CREATE TABLE applimits (id integer primary key default nextval('applimits_seq'),app_id bigint not null  references applicationdefinition (id) on delete cascade  ,maxmemory integer not null  ,priority integer not null  ,maxkillmemory integer not null  ,maxswapmemory integer not null  );

Alter statements:
ALTER TABLE applimits ADD COLUMN IF NOT EXISTS app_id bigint not null references applicationdefinition (id) on delete cascade  default 0;
ALTER TABLE applimits ADD COLUMN IF NOT EXISTS maxmemory integer not null default 0;
ALTER TABLE applimits ADD COLUMN IF NOT EXISTS priority integer not null default 0;
ALTER TABLE applimits ADD COLUMN IF NOT EXISTS maxkillmemory integer not null default 0;
ALTER TABLE applimits ADD COLUMN IF NOT EXISTS maxswapmemory integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE applimits_archive (id integer unique not null,app_id bigint not null,maxmemory integer not null,priority integer not null,maxkillmemory integer not null,maxswapmemory integer not null);
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
	default_def_DBSQLLimits *DBSQLLimits
)

type DBSQLLimits struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBSQLLimits()
	})
}

func DefaultDBSQLLimits() *DBSQLLimits {
	if default_def_DBSQLLimits != nil {
		return default_def_DBSQLLimits
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBSQLLimits(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBSQLLimits = res
	return res
}
func NewDBSQLLimits(db *sql.DB) *DBSQLLimits {
	foo := DBSQLLimits{DB: db}
	foo.SQLTablename = "applimits"
	foo.SQLArchivetablename = "applimits_archive"
	return &foo
}

func (a *DBSQLLimits) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBSQLLimits) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBSQLLimits) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBSQLLimits) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBSQLLimits", "insert into "+a.SQLArchivetablename+" (id,app_id, maxmemory, priority, maxkillmemory, maxswapmemory) values ($1,$2, $3, $4, $5, $6) ", p.ID, p.AppDef.ID, p.MaxMemory, p.Priority, p.MaxKillMemory, p.MaxSwapMemory)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBSQLLimits) buildSaveMap(ctx context.Context, p *savepb.SQLLimits) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["app_id"] = a.get_col_from_proto(p, "app_id")
	res["maxmemory"] = a.get_col_from_proto(p, "maxmemory")
	res["priority"] = a.get_col_from_proto(p, "priority")
	res["maxkillmemory"] = a.get_col_from_proto(p, "maxkillmemory")
	res["maxswapmemory"] = a.get_col_from_proto(p, "maxswapmemory")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBSQLLimits) Save(ctx context.Context, p *savepb.SQLLimits) (uint64, error) {
	qn := "save_DBSQLLimits"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBSQLLimits) SaveWithID(ctx context.Context, p *savepb.SQLLimits) error {
	qn := "insert_DBSQLLimits"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBSQLLimits) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.SQLLimits) (uint64, error) {
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
func (a *DBSQLLimits) SaveOrUpdate(ctx context.Context, p *savepb.SQLLimits) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBSQLLimits) Update(ctx context.Context, p *savepb.SQLLimits) error {
	qn := "DBSQLLimits_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set app_id=$1, maxmemory=$2, priority=$3, maxkillmemory=$4, maxswapmemory=$5 where id = $6", a.get_AppDef_ID(p), a.get_MaxMemory(p), a.get_Priority(p), a.get_MaxKillMemory(p), a.get_MaxSwapMemory(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBSQLLimits) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBSQLLimits_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBSQLLimits) ByID(ctx context.Context, p uint64) (*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No SQLLimits with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) SQLLimits with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBSQLLimits) TryByID(ctx context.Context, p uint64) (*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) SQLLimits with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBSQLLimits) ByIDs(ctx context.Context, p []uint64) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBSQLLimits) All(ctx context.Context) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBSQLLimits" rows with matching AppDef
func (a *DBSQLLimits) ByAppDef(ctx context.Context, p uint64) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByAppDef"
	l, e := a.fromQuery(ctx, qn, "app_id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAppDef: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSQLLimits" rows with multiple matching AppDef
func (a *DBSQLLimits) ByMultiAppDef(ctx context.Context, p []uint64) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByAppDef"
	l, e := a.fromQuery(ctx, qn, "app_id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAppDef: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSQLLimits) ByLikeAppDef(ctx context.Context, p uint64) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByLikeAppDef"
	l, e := a.fromQuery(ctx, qn, "app_id ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAppDef: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSQLLimits" rows with matching MaxMemory
func (a *DBSQLLimits) ByMaxMemory(ctx context.Context, p uint32) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByMaxMemory"
	l, e := a.fromQuery(ctx, qn, "maxmemory = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMaxMemory: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSQLLimits" rows with multiple matching MaxMemory
func (a *DBSQLLimits) ByMultiMaxMemory(ctx context.Context, p []uint32) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByMaxMemory"
	l, e := a.fromQuery(ctx, qn, "maxmemory in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMaxMemory: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSQLLimits) ByLikeMaxMemory(ctx context.Context, p uint32) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByLikeMaxMemory"
	l, e := a.fromQuery(ctx, qn, "maxmemory ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMaxMemory: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSQLLimits" rows with matching Priority
func (a *DBSQLLimits) ByPriority(ctx context.Context, p int32) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByPriority"
	l, e := a.fromQuery(ctx, qn, "priority = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPriority: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSQLLimits" rows with multiple matching Priority
func (a *DBSQLLimits) ByMultiPriority(ctx context.Context, p []int32) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByPriority"
	l, e := a.fromQuery(ctx, qn, "priority in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPriority: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSQLLimits) ByLikePriority(ctx context.Context, p int32) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByLikePriority"
	l, e := a.fromQuery(ctx, qn, "priority ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPriority: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSQLLimits" rows with matching MaxKillMemory
func (a *DBSQLLimits) ByMaxKillMemory(ctx context.Context, p uint32) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByMaxKillMemory"
	l, e := a.fromQuery(ctx, qn, "maxkillmemory = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMaxKillMemory: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSQLLimits" rows with multiple matching MaxKillMemory
func (a *DBSQLLimits) ByMultiMaxKillMemory(ctx context.Context, p []uint32) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByMaxKillMemory"
	l, e := a.fromQuery(ctx, qn, "maxkillmemory in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMaxKillMemory: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSQLLimits) ByLikeMaxKillMemory(ctx context.Context, p uint32) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByLikeMaxKillMemory"
	l, e := a.fromQuery(ctx, qn, "maxkillmemory ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMaxKillMemory: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSQLLimits" rows with matching MaxSwapMemory
func (a *DBSQLLimits) ByMaxSwapMemory(ctx context.Context, p uint32) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByMaxSwapMemory"
	l, e := a.fromQuery(ctx, qn, "maxswapmemory = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMaxSwapMemory: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSQLLimits" rows with multiple matching MaxSwapMemory
func (a *DBSQLLimits) ByMultiMaxSwapMemory(ctx context.Context, p []uint32) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByMaxSwapMemory"
	l, e := a.fromQuery(ctx, qn, "maxswapmemory in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMaxSwapMemory: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSQLLimits) ByLikeMaxSwapMemory(ctx context.Context, p uint32) ([]*savepb.SQLLimits, error) {
	qn := "DBSQLLimits_ByLikeMaxSwapMemory"
	l, e := a.fromQuery(ctx, qn, "maxswapmemory ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMaxSwapMemory: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBSQLLimits) get_ID(p *savepb.SQLLimits) uint64 {
	return uint64(p.ID)
}

// getter for reference "AppDef"
func (a *DBSQLLimits) get_AppDef_ID(p *savepb.SQLLimits) uint64 {
	if p.AppDef == nil {
		panic("field AppDef must not be nil")
	}
	return p.AppDef.ID
}

// getter for field "MaxMemory" (MaxMemory) [uint32]
func (a *DBSQLLimits) get_MaxMemory(p *savepb.SQLLimits) uint32 {
	return uint32(p.MaxMemory)
}

// getter for field "Priority" (Priority) [int32]
func (a *DBSQLLimits) get_Priority(p *savepb.SQLLimits) int32 {
	return int32(p.Priority)
}

// getter for field "MaxKillMemory" (MaxKillMemory) [uint32]
func (a *DBSQLLimits) get_MaxKillMemory(p *savepb.SQLLimits) uint32 {
	return uint32(p.MaxKillMemory)
}

// getter for field "MaxSwapMemory" (MaxSwapMemory) [uint32]
func (a *DBSQLLimits) get_MaxSwapMemory(p *savepb.SQLLimits) uint32 {
	return uint32(p.MaxSwapMemory)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBSQLLimits) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.SQLLimits, error) {
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

func (a *DBSQLLimits) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.SQLLimits, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBSQLLimits) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.SQLLimits, error) {
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
func (a *DBSQLLimits) get_col_from_proto(p *savepb.SQLLimits, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "app_id" {
		return a.get_AppDef_ID(p)
	} else if colname == "maxmemory" {
		return a.get_MaxMemory(p)
	} else if colname == "priority" {
		return a.get_Priority(p)
	} else if colname == "maxkillmemory" {
		return a.get_MaxKillMemory(p)
	} else if colname == "maxswapmemory" {
		return a.get_MaxSwapMemory(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBSQLLimits) Tablename() string {
	return a.SQLTablename
}

func (a *DBSQLLimits) SelectCols() string {
	return "id,app_id, maxmemory, priority, maxkillmemory, maxswapmemory"
}
func (a *DBSQLLimits) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".app_id, " + a.SQLTablename + ".maxmemory, " + a.SQLTablename + ".priority, " + a.SQLTablename + ".maxkillmemory, " + a.SQLTablename + ".maxswapmemory"
}

func (a *DBSQLLimits) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.SQLLimits, error) {
	var res []*savepb.SQLLimits
	for rows.Next() {
		// SCANNER:
		foo := &savepb.SQLLimits{}
		// create the non-nullable pointers
		foo.AppDef = &savepb.ApplicationDefinition{} // non-nullable
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.AppDef.ID
		scanTarget_2 := &foo.MaxMemory
		scanTarget_3 := &foo.Priority
		scanTarget_4 := &foo.MaxKillMemory
		scanTarget_5 := &foo.MaxSwapMemory
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4, scanTarget_5)
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
func (a *DBSQLLimits) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),app_id bigint not null ,maxmemory integer not null ,priority integer not null ,maxkillmemory integer not null ,maxswapmemory integer not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),app_id bigint not null ,maxmemory integer not null ,priority integer not null ,maxkillmemory integer not null ,maxswapmemory integer not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS app_id bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS maxmemory integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS priority integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS maxkillmemory integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS maxswapmemory integer not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS app_id bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS maxmemory integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS priority integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS maxkillmemory integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS maxswapmemory integer not null  default 0;`,
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
		`ALTER TABLE ` + a.SQLTablename + ` add constraint mkdb_fk_applimits_app_id_applicationdefinitionid FOREIGN KEY (app_id) references applicationdefinition (id) on delete cascade ;`,
	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBSQLLimits) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

