package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBApplication
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence deployminator_application_seq;

Main Table:

 CREATE TABLE deployminator_application (id integer primary key default nextval('deployminator_application_seq'),r_binary text not null  ,repositoryid bigint not null  ,downloadurl text not null  );

Alter statements:
ALTER TABLE deployminator_application ADD COLUMN IF NOT EXISTS r_binary text not null default '';
ALTER TABLE deployminator_application ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;
ALTER TABLE deployminator_application ADD COLUMN IF NOT EXISTS downloadurl text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE deployminator_application_archive (id integer unique not null,r_binary text not null,repositoryid bigint not null,downloadurl text not null);
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
	default_def_DBApplication *DBApplication
)

type DBApplication struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBApplication() *DBApplication {
	if default_def_DBApplication != nil {
		return default_def_DBApplication
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBApplication(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBApplication = res
	return res
}
func NewDBApplication(db *sql.DB) *DBApplication {
	foo := DBApplication{DB: db}
	foo.SQLTablename = "deployminator_application"
	foo.SQLArchivetablename = "deployminator_application_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBApplication) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBApplication", "insert into "+a.SQLArchivetablename+" (id,r_binary, repositoryid, downloadurl) values ($1,$2, $3, $4) ", p.ID, p.Binary, p.RepositoryID, p.DownloadURL)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBApplication) Save(ctx context.Context, p *savepb.Application) (uint64, error) {
	qn := "DBApplication_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (r_binary, repositoryid, downloadurl) values ($1, $2, $3) returning id", p.Binary, p.RepositoryID, p.DownloadURL)
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
func (a *DBApplication) SaveWithID(ctx context.Context, p *savepb.Application) error {
	qn := "insert_DBApplication"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,r_binary, repositoryid, downloadurl) values ($1,$2, $3, $4) ", p.ID, p.Binary, p.RepositoryID, p.DownloadURL)
	return a.Error(ctx, qn, e)
}

func (a *DBApplication) Update(ctx context.Context, p *savepb.Application) error {
	qn := "DBApplication_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set r_binary=$1, repositoryid=$2, downloadurl=$3 where id = $4", p.Binary, p.RepositoryID, p.DownloadURL, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBApplication) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBApplication_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBApplication) ByID(ctx context.Context, p uint64) (*savepb.Application, error) {
	qn := "DBApplication_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, repositoryid, downloadurl from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No Application with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Application with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBApplication) TryByID(ctx context.Context, p uint64) (*savepb.Application, error) {
	qn := "DBApplication_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, repositoryid, downloadurl from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Application with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBApplication) All(ctx context.Context) ([]*savepb.Application, error) {
	qn := "DBApplication_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, repositoryid, downloadurl from "+a.SQLTablename+" order by id")
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

// get all "DBApplication" rows with matching Binary
func (a *DBApplication) ByBinary(ctx context.Context, p string) ([]*savepb.Application, error) {
	qn := "DBApplication_ByBinary"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, repositoryid, downloadurl from "+a.SQLTablename+" where r_binary = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBinary: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBinary: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplication) ByLikeBinary(ctx context.Context, p string) ([]*savepb.Application, error) {
	qn := "DBApplication_ByLikeBinary"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, repositoryid, downloadurl from "+a.SQLTablename+" where r_binary ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBinary: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBinary: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplication" rows with matching RepositoryID
func (a *DBApplication) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.Application, error) {
	qn := "DBApplication_ByRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, repositoryid, downloadurl from "+a.SQLTablename+" where repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplication) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.Application, error) {
	qn := "DBApplication_ByLikeRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, repositoryid, downloadurl from "+a.SQLTablename+" where repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBApplication" rows with matching DownloadURL
func (a *DBApplication) ByDownloadURL(ctx context.Context, p string) ([]*savepb.Application, error) {
	qn := "DBApplication_ByDownloadURL"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, repositoryid, downloadurl from "+a.SQLTablename+" where downloadurl = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadURL: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadURL: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBApplication) ByLikeDownloadURL(ctx context.Context, p string) ([]*savepb.Application, error) {
	qn := "DBApplication_ByLikeDownloadURL"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, repositoryid, downloadurl from "+a.SQLTablename+" where downloadurl ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadURL: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDownloadURL: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBApplication) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Application, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBApplication) Tablename() string {
	return a.SQLTablename
}

func (a *DBApplication) SelectCols() string {
	return "id,r_binary, repositoryid, downloadurl"
}
func (a *DBApplication) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".r_binary, " + a.SQLTablename + ".repositoryid, " + a.SQLTablename + ".downloadurl"
}

func (a *DBApplication) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.Application, error) {
	var res []*savepb.Application
	for rows.Next() {
		foo := savepb.Application{}
		err := rows.Scan(&foo.ID, &foo.Binary, &foo.RepositoryID, &foo.DownloadURL)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}
func (a *DBApplication) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.Application, error) {
	var res []*savepb.Application
	for rows.Next() {
		// SCANNER:
		foo := &savepb.Application{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.Binary
		scanTarget_2 := &foo.RepositoryID
		scanTarget_3 := &foo.DownloadURL
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
func (a *DBApplication) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),r_binary text not null ,repositoryid bigint not null ,downloadurl text not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),r_binary text not null ,repositoryid bigint not null ,downloadurl text not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS r_binary text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS downloadurl text not null default '';`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS r_binary text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS repositoryid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS downloadurl text not null  default '';`,
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
func (a *DBApplication) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
