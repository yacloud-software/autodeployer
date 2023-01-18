package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBReplaceRequest
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence deployminator_replacerequest_seq;

Main Table:

 CREATE TABLE deployminator_replacerequest (id integer primary key default nextval('deployminator_replacerequest_seq'),olddeployment bigint not null  references deployminator_deploymentdescriptor (id) on delete cascade  ,newdeployment bigint not null  references deployminator_deploymentdescriptor (id) on delete cascade  );

Alter statements:
ALTER TABLE deployminator_replacerequest ADD COLUMN IF NOT EXISTS olddeployment bigint not null references deployminator_deploymentdescriptor (id) on delete cascade  default 0;
ALTER TABLE deployminator_replacerequest ADD COLUMN IF NOT EXISTS newdeployment bigint not null references deployminator_deploymentdescriptor (id) on delete cascade  default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE deployminator_replacerequest_archive (id integer unique not null,olddeployment bigint not null,newdeployment bigint not null);
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
	default_def_DBReplaceRequest *DBReplaceRequest
)

type DBReplaceRequest struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBReplaceRequest() *DBReplaceRequest {
	if default_def_DBReplaceRequest != nil {
		return default_def_DBReplaceRequest
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBReplaceRequest(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBReplaceRequest = res
	return res
}
func NewDBReplaceRequest(db *sql.DB) *DBReplaceRequest {
	foo := DBReplaceRequest{DB: db}
	foo.SQLTablename = "deployminator_replacerequest"
	foo.SQLArchivetablename = "deployminator_replacerequest_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBReplaceRequest) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBReplaceRequest", "insert into "+a.SQLArchivetablename+" (id,olddeployment, newdeployment) values ($1,$2, $3) ", p.ID, p.OldDeployment.ID, p.NewDeployment.ID)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBReplaceRequest) Save(ctx context.Context, p *savepb.ReplaceRequest) (uint64, error) {
	qn := "DBReplaceRequest_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (olddeployment, newdeployment) values ($1, $2) returning id", p.OldDeployment.ID, p.NewDeployment.ID)
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
func (a *DBReplaceRequest) SaveWithID(ctx context.Context, p *savepb.ReplaceRequest) error {
	qn := "insert_DBReplaceRequest"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,olddeployment, newdeployment) values ($1,$2, $3) ", p.ID, p.OldDeployment.ID, p.NewDeployment.ID)
	return a.Error(ctx, qn, e)
}

func (a *DBReplaceRequest) Update(ctx context.Context, p *savepb.ReplaceRequest) error {
	qn := "DBReplaceRequest_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set olddeployment=$1, newdeployment=$2 where id = $3", p.OldDeployment.ID, p.NewDeployment.ID, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBReplaceRequest) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBReplaceRequest_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBReplaceRequest) ByID(ctx context.Context, p uint64) (*savepb.ReplaceRequest, error) {
	qn := "DBReplaceRequest_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,olddeployment, newdeployment from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No ReplaceRequest with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) ReplaceRequest with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBReplaceRequest) TryByID(ctx context.Context, p uint64) (*savepb.ReplaceRequest, error) {
	qn := "DBReplaceRequest_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,olddeployment, newdeployment from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) ReplaceRequest with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBReplaceRequest) All(ctx context.Context) ([]*savepb.ReplaceRequest, error) {
	qn := "DBReplaceRequest_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,olddeployment, newdeployment from "+a.SQLTablename+" order by id")
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

// get all "DBReplaceRequest" rows with matching OldDeployment
func (a *DBReplaceRequest) ByOldDeployment(ctx context.Context, p uint64) ([]*savepb.ReplaceRequest, error) {
	qn := "DBReplaceRequest_ByOldDeployment"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,olddeployment, newdeployment from "+a.SQLTablename+" where olddeployment = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOldDeployment: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOldDeployment: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBReplaceRequest) ByLikeOldDeployment(ctx context.Context, p uint64) ([]*savepb.ReplaceRequest, error) {
	qn := "DBReplaceRequest_ByLikeOldDeployment"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,olddeployment, newdeployment from "+a.SQLTablename+" where olddeployment ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOldDeployment: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOldDeployment: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBReplaceRequest" rows with matching NewDeployment
func (a *DBReplaceRequest) ByNewDeployment(ctx context.Context, p uint64) ([]*savepb.ReplaceRequest, error) {
	qn := "DBReplaceRequest_ByNewDeployment"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,olddeployment, newdeployment from "+a.SQLTablename+" where newdeployment = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNewDeployment: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNewDeployment: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBReplaceRequest) ByLikeNewDeployment(ctx context.Context, p uint64) ([]*savepb.ReplaceRequest, error) {
	qn := "DBReplaceRequest_ByLikeNewDeployment"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,olddeployment, newdeployment from "+a.SQLTablename+" where newdeployment ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNewDeployment: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNewDeployment: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBReplaceRequest) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.ReplaceRequest, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBReplaceRequest) Tablename() string {
	return a.SQLTablename
}

func (a *DBReplaceRequest) SelectCols() string {
	return "id,olddeployment, newdeployment"
}
func (a *DBReplaceRequest) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".olddeployment, " + a.SQLTablename + ".newdeployment"
}

func (a *DBReplaceRequest) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.ReplaceRequest, error) {
	var res []*savepb.ReplaceRequest
	for rows.Next() {
		foo := savepb.ReplaceRequest{OldDeployment: &savepb.DeploymentDescriptor{}, NewDeployment: &savepb.DeploymentDescriptor{}}
		err := rows.Scan(&foo.ID, &foo.OldDeployment.ID, &foo.NewDeployment.ID)
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
func (a *DBReplaceRequest) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),olddeployment bigint not null  references deployminator_deploymentdescriptor (id) on delete cascade  ,newdeployment bigint not null  references deployminator_deploymentdescriptor (id) on delete cascade  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),olddeployment bigint not null  references deployminator_deploymentdescriptor (id) on delete cascade  ,newdeployment bigint not null  references deployminator_deploymentdescriptor (id) on delete cascade  );`,
		`ALTER TABLE deployminator_replacerequest ADD COLUMN IF NOT EXISTS olddeployment bigint not null references deployminator_deploymentdescriptor (id) on delete cascade  default 0;`,
		`ALTER TABLE deployminator_replacerequest ADD COLUMN IF NOT EXISTS newdeployment bigint not null references deployminator_deploymentdescriptor (id) on delete cascade  default 0;`,
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
func (a *DBReplaceRequest) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
