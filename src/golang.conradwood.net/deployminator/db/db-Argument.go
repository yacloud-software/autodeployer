package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBArgument
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence deployminator_argument_seq;

Main Table:

 CREATE TABLE deployminator_argument (id integer primary key default nextval('deployminator_argument_seq'),instancedef bigint not null  references deployminator_instancedef (id) on delete cascade  ,argument text not null  );

Alter statements:
ALTER TABLE deployminator_argument ADD COLUMN IF NOT EXISTS instancedef bigint not null references deployminator_instancedef (id) on delete cascade  default 0;
ALTER TABLE deployminator_argument ADD COLUMN IF NOT EXISTS argument text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE deployminator_argument_archive (id integer unique not null,instancedef bigint not null,argument text not null);
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
	default_def_DBArgument *DBArgument
)

type DBArgument struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBArgument() *DBArgument {
	if default_def_DBArgument != nil {
		return default_def_DBArgument
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBArgument(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBArgument = res
	return res
}
func NewDBArgument(db *sql.DB) *DBArgument {
	foo := DBArgument{DB: db}
	foo.SQLTablename = "deployminator_argument"
	foo.SQLArchivetablename = "deployminator_argument_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBArgument) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBArgument", "insert into "+a.SQLArchivetablename+" (id,instancedef, argument) values ($1,$2, $3) ", p.ID, p.InstanceDef.ID, p.Argument)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBArgument) Save(ctx context.Context, p *savepb.Argument) (uint64, error) {
	qn := "DBArgument_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (instancedef, argument) values ($1, $2) returning id", p.InstanceDef.ID, p.Argument)
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
func (a *DBArgument) SaveWithID(ctx context.Context, p *savepb.Argument) error {
	qn := "insert_DBArgument"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,instancedef, argument) values ($1,$2, $3) ", p.ID, p.InstanceDef.ID, p.Argument)
	return a.Error(ctx, qn, e)
}

func (a *DBArgument) Update(ctx context.Context, p *savepb.Argument) error {
	qn := "DBArgument_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set instancedef=$1, argument=$2 where id = $3", p.InstanceDef.ID, p.Argument, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBArgument) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBArgument_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBArgument) ByID(ctx context.Context, p uint64) (*savepb.Argument, error) {
	qn := "DBArgument_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,instancedef, argument from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No Argument with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Argument with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBArgument) TryByID(ctx context.Context, p uint64) (*savepb.Argument, error) {
	qn := "DBArgument_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,instancedef, argument from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Argument with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBArgument) All(ctx context.Context) ([]*savepb.Argument, error) {
	qn := "DBArgument_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,instancedef, argument from "+a.SQLTablename+" order by id")
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

// get all "DBArgument" rows with matching InstanceDef
func (a *DBArgument) ByInstanceDef(ctx context.Context, p uint64) ([]*savepb.Argument, error) {
	qn := "DBArgument_ByInstanceDef"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,instancedef, argument from "+a.SQLTablename+" where instancedef = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInstanceDef: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInstanceDef: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBArgument) ByLikeInstanceDef(ctx context.Context, p uint64) ([]*savepb.Argument, error) {
	qn := "DBArgument_ByLikeInstanceDef"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,instancedef, argument from "+a.SQLTablename+" where instancedef ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInstanceDef: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInstanceDef: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBArgument" rows with matching Argument
func (a *DBArgument) ByArgument(ctx context.Context, p string) ([]*savepb.Argument, error) {
	qn := "DBArgument_ByArgument"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,instancedef, argument from "+a.SQLTablename+" where argument = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByArgument: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByArgument: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBArgument) ByLikeArgument(ctx context.Context, p string) ([]*savepb.Argument, error) {
	qn := "DBArgument_ByLikeArgument"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,instancedef, argument from "+a.SQLTablename+" where argument ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByArgument: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByArgument: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBArgument) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Argument, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBArgument) Tablename() string {
	return a.SQLTablename
}

func (a *DBArgument) SelectCols() string {
	return "id,instancedef, argument"
}
func (a *DBArgument) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".instancedef, " + a.SQLTablename + ".argument"
}

func (a *DBArgument) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.Argument, error) {
	var res []*savepb.Argument
	for rows.Next() {
		foo := savepb.Argument{InstanceDef: &savepb.InstanceDef{}}
		err := rows.Scan(&foo.ID, &foo.InstanceDef.ID, &foo.Argument)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}
func (a *DBArgument) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.Argument, error) {
	var res []*savepb.Argument
	for rows.Next() {
		// SCANNER:
		foo := &savepb.Argument{}
		// create the non-nullable pointers
		foo.InstanceDef = &savepb.InstanceDef{} // non-nullable
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.InstanceDef.ID
		scanTarget_2 := &foo.Argument
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
func (a *DBArgument) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),instancedef bigint not null ,argument text not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),instancedef bigint not null ,argument text not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS instancedef bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS argument text not null default '';`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS instancedef bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS argument text not null  default '';`,
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
		`ALTER TABLE deployminator_argument add constraint mkdb_fk_deployminator_argument_instancedef_deployminator_instancedefid FOREIGN KEY (instancedef) references deployminator_instancedef (id) on delete cascade ;`,
	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBArgument) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

