package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBDeploymentLog
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence deploymentlog_seq;

Main Table:

 CREATE TABLE deploymentlog (id integer primary key default nextval('deploymentlog_seq'),r_binary text not null  ,app_id bigint not null  references applicationdefinition (id) on delete cascade  ,buildid bigint not null  ,autodeployerhost text not null  ,started integer not null  ,finished integer not null  ,message text not null  ,deployalgorithm integer not null  );

Alter statements:
ALTER TABLE deploymentlog ADD COLUMN IF NOT EXISTS r_binary text not null default '';
ALTER TABLE deploymentlog ADD COLUMN IF NOT EXISTS app_id bigint not null references applicationdefinition (id) on delete cascade  default 0;
ALTER TABLE deploymentlog ADD COLUMN IF NOT EXISTS buildid bigint not null default 0;
ALTER TABLE deploymentlog ADD COLUMN IF NOT EXISTS autodeployerhost text not null default '';
ALTER TABLE deploymentlog ADD COLUMN IF NOT EXISTS started integer not null default 0;
ALTER TABLE deploymentlog ADD COLUMN IF NOT EXISTS finished integer not null default 0;
ALTER TABLE deploymentlog ADD COLUMN IF NOT EXISTS message text not null default '';
ALTER TABLE deploymentlog ADD COLUMN IF NOT EXISTS deployalgorithm integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE deploymentlog_archive (id integer unique not null,r_binary text not null,app_id bigint not null,buildid bigint not null,autodeployerhost text not null,started integer not null,finished integer not null,message text not null,deployalgorithm integer not null);
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
	default_def_DBDeploymentLog *DBDeploymentLog
)

type DBDeploymentLog struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBDeploymentLog() *DBDeploymentLog {
	if default_def_DBDeploymentLog != nil {
		return default_def_DBDeploymentLog
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBDeploymentLog(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBDeploymentLog = res
	return res
}
func NewDBDeploymentLog(db *sql.DB) *DBDeploymentLog {
	foo := DBDeploymentLog{DB: db}
	foo.SQLTablename = "deploymentlog"
	foo.SQLArchivetablename = "deploymentlog_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBDeploymentLog) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBDeploymentLog", "insert into "+a.SQLArchivetablename+" (id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm) values ($1,$2, $3, $4, $5, $6, $7, $8, $9) ", p.ID, p.Binary, p.AppDef.ID, p.BuildID, p.AutoDeployerHost, p.Started, p.Finished, p.Message, p.DeployAlgorithm)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBDeploymentLog) Save(ctx context.Context, p *savepb.DeploymentLog) (uint64, error) {
	qn := "DBDeploymentLog_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id", a.get_Binary(p), a.get_AppDef_ID(p), a.get_BuildID(p), a.get_AutoDeployerHost(p), a.get_Started(p), a.get_Finished(p), a.get_Message(p), a.get_DeployAlgorithm(p))
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
func (a *DBDeploymentLog) SaveWithID(ctx context.Context, p *savepb.DeploymentLog) error {
	qn := "insert_DBDeploymentLog"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm) values ($1,$2, $3, $4, $5, $6, $7, $8, $9) ", p.ID, p.Binary, p.AppDef.ID, p.BuildID, p.AutoDeployerHost, p.Started, p.Finished, p.Message, p.DeployAlgorithm)
	return a.Error(ctx, qn, e)
}

func (a *DBDeploymentLog) Update(ctx context.Context, p *savepb.DeploymentLog) error {
	qn := "DBDeploymentLog_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set r_binary=$1, app_id=$2, buildid=$3, autodeployerhost=$4, started=$5, finished=$6, message=$7, deployalgorithm=$8 where id = $9", a.get_Binary(p), a.get_AppDef_ID(p), a.get_BuildID(p), a.get_AutoDeployerHost(p), a.get_Started(p), a.get_Finished(p), a.get_Message(p), a.get_DeployAlgorithm(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBDeploymentLog) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBDeploymentLog_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBDeploymentLog) ByID(ctx context.Context, p uint64) (*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No DeploymentLog with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) DeploymentLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBDeploymentLog) TryByID(ctx context.Context, p uint64) (*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) DeploymentLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBDeploymentLog) All(ctx context.Context) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" order by id")
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

// get all "DBDeploymentLog" rows with matching Binary
func (a *DBDeploymentLog) ByBinary(ctx context.Context, p string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByBinary"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where r_binary = $1", p)
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
func (a *DBDeploymentLog) ByLikeBinary(ctx context.Context, p string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeBinary"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where r_binary ilike $1", p)
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

// get all "DBDeploymentLog" rows with matching AppDef
func (a *DBDeploymentLog) ByAppDef(ctx context.Context, p uint64) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByAppDef"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where app_id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAppDef: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAppDef: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeAppDef(ctx context.Context, p uint64) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeAppDef"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where app_id ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAppDef: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAppDef: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with matching BuildID
func (a *DBDeploymentLog) ByBuildID(ctx context.Context, p uint64) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByBuildID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where buildid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBuildID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBuildID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeBuildID(ctx context.Context, p uint64) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeBuildID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where buildid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBuildID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBuildID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with matching AutoDeployerHost
func (a *DBDeploymentLog) ByAutoDeployerHost(ctx context.Context, p string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByAutoDeployerHost"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where autodeployerhost = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAutoDeployerHost: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAutoDeployerHost: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeAutoDeployerHost(ctx context.Context, p string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeAutoDeployerHost"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where autodeployerhost ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAutoDeployerHost: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAutoDeployerHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with matching Started
func (a *DBDeploymentLog) ByStarted(ctx context.Context, p uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByStarted"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where started = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByStarted: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByStarted: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeStarted(ctx context.Context, p uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeStarted"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where started ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByStarted: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByStarted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with matching Finished
func (a *DBDeploymentLog) ByFinished(ctx context.Context, p uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByFinished"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where finished = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFinished: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFinished: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeFinished(ctx context.Context, p uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeFinished"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where finished ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFinished: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFinished: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with matching Message
func (a *DBDeploymentLog) ByMessage(ctx context.Context, p string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByMessage"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where message = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMessage: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMessage: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeMessage(ctx context.Context, p string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeMessage"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where message ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMessage: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with matching DeployAlgorithm
func (a *DBDeploymentLog) ByDeployAlgorithm(ctx context.Context, p uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByDeployAlgorithm"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where deployalgorithm = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployAlgorithm: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployAlgorithm: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeDeployAlgorithm(ctx context.Context, p uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeDeployAlgorithm"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm from "+a.SQLTablename+" where deployalgorithm ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployAlgorithm: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeployAlgorithm: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBDeploymentLog) get_ID(p *savepb.DeploymentLog) uint64 {
	return uint64(p.ID)
}

// getter for field "Binary" (Binary) [string]
func (a *DBDeploymentLog) get_Binary(p *savepb.DeploymentLog) string {
	return string(p.Binary)
}

// getter for reference "AppDef"
func (a *DBDeploymentLog) get_AppDef_ID(p *savepb.DeploymentLog) uint64 {
	if p.AppDef == nil {
		panic("field AppDef must not be nil")
	}
	return p.AppDef.ID
}

// getter for field "BuildID" (BuildID) [uint64]
func (a *DBDeploymentLog) get_BuildID(p *savepb.DeploymentLog) uint64 {
	return uint64(p.BuildID)
}

// getter for field "AutoDeployerHost" (AutoDeployerHost) [string]
func (a *DBDeploymentLog) get_AutoDeployerHost(p *savepb.DeploymentLog) string {
	return string(p.AutoDeployerHost)
}

// getter for field "Started" (Started) [uint32]
func (a *DBDeploymentLog) get_Started(p *savepb.DeploymentLog) uint32 {
	return uint32(p.Started)
}

// getter for field "Finished" (Finished) [uint32]
func (a *DBDeploymentLog) get_Finished(p *savepb.DeploymentLog) uint32 {
	return uint32(p.Finished)
}

// getter for field "Message" (Message) [string]
func (a *DBDeploymentLog) get_Message(p *savepb.DeploymentLog) string {
	return string(p.Message)
}

// getter for field "DeployAlgorithm" (DeployAlgorithm) [uint32]
func (a *DBDeploymentLog) get_DeployAlgorithm(p *savepb.DeploymentLog) uint32 {
	return uint32(p.DeployAlgorithm)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBDeploymentLog) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.DeploymentLog, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBDeploymentLog) Tablename() string {
	return a.SQLTablename
}

func (a *DBDeploymentLog) SelectCols() string {
	return "id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm"
}
func (a *DBDeploymentLog) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".r_binary, " + a.SQLTablename + ".app_id, " + a.SQLTablename + ".buildid, " + a.SQLTablename + ".autodeployerhost, " + a.SQLTablename + ".started, " + a.SQLTablename + ".finished, " + a.SQLTablename + ".message, " + a.SQLTablename + ".deployalgorithm"
}

func (a *DBDeploymentLog) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.DeploymentLog, error) {
	var res []*savepb.DeploymentLog
	for rows.Next() {
		foo := savepb.DeploymentLog{AppDef: &savepb.ApplicationDefinition{}}
		err := rows.Scan(&foo.ID, &foo.Binary, &foo.AppDef.ID, &foo.BuildID, &foo.AutoDeployerHost, &foo.Started, &foo.Finished, &foo.Message, &foo.DeployAlgorithm)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}
func (a *DBDeploymentLog) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.DeploymentLog, error) {
	var res []*savepb.DeploymentLog
	for rows.Next() {
		// SCANNER:
		foo := &savepb.DeploymentLog{}
		// create the non-nullable pointers
		foo.AppDef = &savepb.ApplicationDefinition{} // non-nullable
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.Binary
		scanTarget_2 := &foo.AppDef.ID
		scanTarget_3 := &foo.BuildID
		scanTarget_4 := &foo.AutoDeployerHost
		scanTarget_5 := &foo.Started
		scanTarget_6 := &foo.Finished
		scanTarget_7 := &foo.Message
		scanTarget_8 := &foo.DeployAlgorithm
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4, scanTarget_5, scanTarget_6, scanTarget_7, scanTarget_8)
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
func (a *DBDeploymentLog) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),r_binary text not null ,app_id bigint not null ,buildid bigint not null ,autodeployerhost text not null ,started integer not null ,finished integer not null ,message text not null ,deployalgorithm integer not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),r_binary text not null ,app_id bigint not null ,buildid bigint not null ,autodeployerhost text not null ,started integer not null ,finished integer not null ,message text not null ,deployalgorithm integer not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS r_binary text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS app_id bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS buildid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS autodeployerhost text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS started integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS finished integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS message text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS deployalgorithm integer not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS r_binary text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS app_id bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS buildid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS autodeployerhost text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS started integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS finished integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS message text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS deployalgorithm integer not null  default 0;`,
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
		`ALTER TABLE ` + a.SQLTablename + ` add constraint mkdb_fk_deploymentlog_app_id_applicationdefinitionid FOREIGN KEY (app_id) references applicationdefinition (id) on delete cascade ;`,
	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBDeploymentLog) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

