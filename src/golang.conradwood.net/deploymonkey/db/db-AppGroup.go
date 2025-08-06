package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBAppGroup
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
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/sql"
	"os"
	"sync"
)

var (
	default_def_DBAppGroup *DBAppGroup
)

type DBAppGroup struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBAppGroup()
	})
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

func (a *DBAppGroup) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBAppGroup) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBAppGroup) NewQuery() *Query {
	return newQuery(a)
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

// return a map with columnname -> value_from_proto
func (a *DBAppGroup) buildSaveMap(ctx context.Context, p *savepb.AppGroup) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["namespace"] = a.get_col_from_proto(p, "namespace")
	res["groupname"] = a.get_col_from_proto(p, "groupname")
	res["deployedversion"] = a.get_col_from_proto(p, "deployedversion")
	res["pendingversion"] = a.get_col_from_proto(p, "pendingversion")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBAppGroup) Save(ctx context.Context, p *savepb.AppGroup) (uint64, error) {
	qn := "save_DBAppGroup"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBAppGroup) SaveWithID(ctx context.Context, p *savepb.AppGroup) error {
	qn := "insert_DBAppGroup"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBAppGroup) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.AppGroup) (uint64, error) {
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
func (a *DBAppGroup) SaveOrUpdate(ctx context.Context, p *savepb.AppGroup) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
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
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No AppGroup with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) AppGroup with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBAppGroup) TryByID(ctx context.Context, p uint64) (*savepb.AppGroup, error) {
	qn := "DBAppGroup_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) AppGroup with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBAppGroup) ByIDs(ctx context.Context, p []uint64) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBAppGroup) All(ctx context.Context) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBAppGroup" rows with matching Namespace
func (a *DBAppGroup) ByNamespace(ctx context.Context, p string) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByNamespace"
	l, e := a.fromQuery(ctx, qn, "namespace = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByNamespace: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppGroup" rows with multiple matching Namespace
func (a *DBAppGroup) ByMultiNamespace(ctx context.Context, p []string) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByNamespace"
	l, e := a.fromQuery(ctx, qn, "namespace in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByNamespace: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBAppGroup) ByLikeNamespace(ctx context.Context, p string) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByLikeNamespace"
	l, e := a.fromQuery(ctx, qn, "namespace ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByNamespace: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppGroup" rows with matching Groupname
func (a *DBAppGroup) ByGroupname(ctx context.Context, p string) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByGroupname"
	l, e := a.fromQuery(ctx, qn, "groupname = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByGroupname: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppGroup" rows with multiple matching Groupname
func (a *DBAppGroup) ByMultiGroupname(ctx context.Context, p []string) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByGroupname"
	l, e := a.fromQuery(ctx, qn, "groupname in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByGroupname: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBAppGroup) ByLikeGroupname(ctx context.Context, p string) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByLikeGroupname"
	l, e := a.fromQuery(ctx, qn, "groupname ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByGroupname: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppGroup" rows with matching DeployedVersion
func (a *DBAppGroup) ByDeployedVersion(ctx context.Context, p uint32) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByDeployedVersion"
	l, e := a.fromQuery(ctx, qn, "deployedversion = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeployedVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppGroup" rows with multiple matching DeployedVersion
func (a *DBAppGroup) ByMultiDeployedVersion(ctx context.Context, p []uint32) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByDeployedVersion"
	l, e := a.fromQuery(ctx, qn, "deployedversion in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeployedVersion: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBAppGroup) ByLikeDeployedVersion(ctx context.Context, p uint32) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByLikeDeployedVersion"
	l, e := a.fromQuery(ctx, qn, "deployedversion ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeployedVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppGroup" rows with matching PendingVersion
func (a *DBAppGroup) ByPendingVersion(ctx context.Context, p uint32) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByPendingVersion"
	l, e := a.fromQuery(ctx, qn, "pendingversion = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPendingVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBAppGroup" rows with multiple matching PendingVersion
func (a *DBAppGroup) ByMultiPendingVersion(ctx context.Context, p []uint32) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByPendingVersion"
	l, e := a.fromQuery(ctx, qn, "pendingversion in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPendingVersion: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBAppGroup) ByLikePendingVersion(ctx context.Context, p uint32) ([]*savepb.AppGroup, error) {
	qn := "DBAppGroup_ByLikePendingVersion"
	l, e := a.fromQuery(ctx, qn, "pendingversion ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPendingVersion: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBAppGroup) get_ID(p *savepb.AppGroup) uint64 {
	return uint64(p.ID)
}

// getter for field "Namespace" (Namespace) [string]
func (a *DBAppGroup) get_Namespace(p *savepb.AppGroup) string {
	return string(p.Namespace)
}

// getter for field "Groupname" (Groupname) [string]
func (a *DBAppGroup) get_Groupname(p *savepb.AppGroup) string {
	return string(p.Groupname)
}

// getter for field "DeployedVersion" (DeployedVersion) [uint32]
func (a *DBAppGroup) get_DeployedVersion(p *savepb.AppGroup) uint32 {
	return uint32(p.DeployedVersion)
}

// getter for field "PendingVersion" (PendingVersion) [uint32]
func (a *DBAppGroup) get_PendingVersion(p *savepb.AppGroup) uint32 {
	return uint32(p.PendingVersion)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBAppGroup) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.AppGroup, error) {
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

func (a *DBAppGroup) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.AppGroup, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBAppGroup) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.AppGroup, error) {
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
func (a *DBAppGroup) get_col_from_proto(p *savepb.AppGroup, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "namespace" {
		return a.get_Namespace(p)
	} else if colname == "groupname" {
		return a.get_Groupname(p)
	} else if colname == "deployedversion" {
		return a.get_DeployedVersion(p)
	} else if colname == "pendingversion" {
		return a.get_PendingVersion(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBAppGroup) Tablename() string {
	return a.SQLTablename
}

func (a *DBAppGroup) SelectCols() string {
	return "id,namespace, groupname, deployedversion, pendingversion"
}
func (a *DBAppGroup) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".namespace, " + a.SQLTablename + ".groupname, " + a.SQLTablename + ".deployedversion, " + a.SQLTablename + ".pendingversion"
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
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

