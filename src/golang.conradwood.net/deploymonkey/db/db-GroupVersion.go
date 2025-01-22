package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBGroupVersion
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence groupversion_seq;

Main Table:

 CREATE TABLE groupversion (id integer primary key default nextval('groupversion_seq'),group_id bigint not null  references appgroup (id) on delete cascade  ,createdtimestamp integer not null  );

Alter statements:
ALTER TABLE groupversion ADD COLUMN IF NOT EXISTS group_id bigint not null references appgroup (id) on delete cascade  default 0;
ALTER TABLE groupversion ADD COLUMN IF NOT EXISTS createdtimestamp integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE groupversion_archive (id integer unique not null,group_id bigint not null,createdtimestamp integer not null);
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
	default_def_DBGroupVersion *DBGroupVersion
)

type DBGroupVersion struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func DefaultDBGroupVersion() *DBGroupVersion {
	if default_def_DBGroupVersion != nil {
		return default_def_DBGroupVersion
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBGroupVersion(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBGroupVersion = res
	return res
}
func NewDBGroupVersion(db *sql.DB) *DBGroupVersion {
	foo := DBGroupVersion{DB: db}
	foo.SQLTablename = "groupversion"
	foo.SQLArchivetablename = "groupversion_archive"
	return &foo
}

func (a *DBGroupVersion) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBGroupVersion) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

// archive. It is NOT transactionally save.
func (a *DBGroupVersion) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBGroupVersion", "insert into "+a.SQLArchivetablename+" (id,group_id, createdtimestamp) values ($1,$2, $3) ", p.ID, p.GroupID.ID, p.CreatedTimestamp)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBGroupVersion) buildSaveMap(ctx context.Context, p *savepb.GroupVersion) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["group_id"] = a.get_col_from_proto(p, "group_id")
	res["createdtimestamp"] = a.get_col_from_proto(p, "createdtimestamp")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBGroupVersion) Save(ctx context.Context, p *savepb.GroupVersion) (uint64, error) {
	qn := "save_DBGroupVersion"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBGroupVersion) SaveWithID(ctx context.Context, p *savepb.GroupVersion) error {
	qn := "insert_DBGroupVersion"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBGroupVersion) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.GroupVersion) (uint64, error) {
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

func (a *DBGroupVersion) Update(ctx context.Context, p *savepb.GroupVersion) error {
	qn := "DBGroupVersion_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set group_id=$1, createdtimestamp=$2 where id = $3", a.get_GroupID_ID(p), a.get_CreatedTimestamp(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBGroupVersion) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBGroupVersion_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBGroupVersion) ByID(ctx context.Context, p uint64) (*savepb.GroupVersion, error) {
	qn := "DBGroupVersion_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No GroupVersion with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) GroupVersion with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBGroupVersion) TryByID(ctx context.Context, p uint64) (*savepb.GroupVersion, error) {
	qn := "DBGroupVersion_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) GroupVersion with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBGroupVersion) ByIDs(ctx context.Context, p []uint64) ([]*savepb.GroupVersion, error) {
	qn := "DBGroupVersion_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBGroupVersion) All(ctx context.Context) ([]*savepb.GroupVersion, error) {
	qn := "DBGroupVersion_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBGroupVersion" rows with matching GroupID
func (a *DBGroupVersion) ByGroupID(ctx context.Context, p uint64) ([]*savepb.GroupVersion, error) {
	qn := "DBGroupVersion_ByGroupID"
	l, e := a.fromQuery(ctx, qn, "group_id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByGroupID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGroupVersion" rows with multiple matching GroupID
func (a *DBGroupVersion) ByMultiGroupID(ctx context.Context, p []uint64) ([]*savepb.GroupVersion, error) {
	qn := "DBGroupVersion_ByGroupID"
	l, e := a.fromQuery(ctx, qn, "group_id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByGroupID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGroupVersion) ByLikeGroupID(ctx context.Context, p uint64) ([]*savepb.GroupVersion, error) {
	qn := "DBGroupVersion_ByLikeGroupID"
	l, e := a.fromQuery(ctx, qn, "group_id ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByGroupID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGroupVersion" rows with matching CreatedTimestamp
func (a *DBGroupVersion) ByCreatedTimestamp(ctx context.Context, p uint32) ([]*savepb.GroupVersion, error) {
	qn := "DBGroupVersion_ByCreatedTimestamp"
	l, e := a.fromQuery(ctx, qn, "createdtimestamp = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreatedTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGroupVersion" rows with multiple matching CreatedTimestamp
func (a *DBGroupVersion) ByMultiCreatedTimestamp(ctx context.Context, p []uint32) ([]*savepb.GroupVersion, error) {
	qn := "DBGroupVersion_ByCreatedTimestamp"
	l, e := a.fromQuery(ctx, qn, "createdtimestamp in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreatedTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGroupVersion) ByLikeCreatedTimestamp(ctx context.Context, p uint32) ([]*savepb.GroupVersion, error) {
	qn := "DBGroupVersion_ByLikeCreatedTimestamp"
	l, e := a.fromQuery(ctx, qn, "createdtimestamp ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreatedTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBGroupVersion) get_ID(p *savepb.GroupVersion) uint64 {
	return uint64(p.ID)
}

// getter for reference "GroupID"
func (a *DBGroupVersion) get_GroupID_ID(p *savepb.GroupVersion) uint64 {
	if p.GroupID == nil {
		panic("field GroupID must not be nil")
	}
	return p.GroupID.ID
}

// getter for field "CreatedTimestamp" (CreatedTimestamp) [uint32]
func (a *DBGroupVersion) get_CreatedTimestamp(p *savepb.GroupVersion) uint32 {
	return uint32(p.CreatedTimestamp)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBGroupVersion) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.GroupVersion, error) {
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

func (a *DBGroupVersion) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.GroupVersion, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBGroupVersion) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.GroupVersion, error) {
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
func (a *DBGroupVersion) get_col_from_proto(p *savepb.GroupVersion, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "group_id" {
		return a.get_GroupID_ID(p)
	} else if colname == "createdtimestamp" {
		return a.get_CreatedTimestamp(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBGroupVersion) Tablename() string {
	return a.SQLTablename
}

func (a *DBGroupVersion) SelectCols() string {
	return "id,group_id, createdtimestamp"
}
func (a *DBGroupVersion) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".group_id, " + a.SQLTablename + ".createdtimestamp"
}

func (a *DBGroupVersion) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.GroupVersion, error) {
	var res []*savepb.GroupVersion
	for rows.Next() {
		// SCANNER:
		foo := &savepb.GroupVersion{}
		// create the non-nullable pointers
		foo.GroupID = &savepb.AppGroup{} // non-nullable
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.GroupID.ID
		scanTarget_2 := &foo.CreatedTimestamp
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
func (a *DBGroupVersion) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),group_id bigint not null ,createdtimestamp integer not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),group_id bigint not null ,createdtimestamp integer not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS group_id bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS createdtimestamp integer not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS group_id bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS createdtimestamp integer not null  default 0;`,
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
		`ALTER TABLE ` + a.SQLTablename + ` add constraint mkdb_fk_groupversion_group_id_appgroupid FOREIGN KEY (group_id) references appgroup (id) on delete cascade ;`,
	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBGroupVersion) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

