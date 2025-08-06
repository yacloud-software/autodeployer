package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBDeploymentLog
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
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/sql"
	"os"
	"sync"
)

var (
	default_def_DBDeploymentLog *DBDeploymentLog
)

type DBDeploymentLog struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBDeploymentLog()
	})
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

func (a *DBDeploymentLog) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBDeploymentLog) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBDeploymentLog) NewQuery() *Query {
	return newQuery(a)
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

// return a map with columnname -> value_from_proto
func (a *DBDeploymentLog) buildSaveMap(ctx context.Context, p *savepb.DeploymentLog) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["r_binary"] = a.get_col_from_proto(p, "r_binary")
	res["app_id"] = a.get_col_from_proto(p, "app_id")
	res["buildid"] = a.get_col_from_proto(p, "buildid")
	res["autodeployerhost"] = a.get_col_from_proto(p, "autodeployerhost")
	res["started"] = a.get_col_from_proto(p, "started")
	res["finished"] = a.get_col_from_proto(p, "finished")
	res["message"] = a.get_col_from_proto(p, "message")
	res["deployalgorithm"] = a.get_col_from_proto(p, "deployalgorithm")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBDeploymentLog) Save(ctx context.Context, p *savepb.DeploymentLog) (uint64, error) {
	qn := "save_DBDeploymentLog"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBDeploymentLog) SaveWithID(ctx context.Context, p *savepb.DeploymentLog) error {
	qn := "insert_DBDeploymentLog"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBDeploymentLog) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.DeploymentLog) (uint64, error) {
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
func (a *DBDeploymentLog) SaveOrUpdate(ctx context.Context, p *savepb.DeploymentLog) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
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
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No DeploymentLog with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) DeploymentLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBDeploymentLog) TryByID(ctx context.Context, p uint64) (*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) DeploymentLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBDeploymentLog) ByIDs(ctx context.Context, p []uint64) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBDeploymentLog) All(ctx context.Context) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBDeploymentLog" rows with matching Binary
func (a *DBDeploymentLog) ByBinary(ctx context.Context, p string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByBinary"
	l, e := a.fromQuery(ctx, qn, "r_binary = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBinary: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with multiple matching Binary
func (a *DBDeploymentLog) ByMultiBinary(ctx context.Context, p []string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByBinary"
	l, e := a.fromQuery(ctx, qn, "r_binary in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBinary: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeBinary(ctx context.Context, p string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeBinary"
	l, e := a.fromQuery(ctx, qn, "r_binary ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBinary: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with matching AppDef
func (a *DBDeploymentLog) ByAppDef(ctx context.Context, p uint64) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByAppDef"
	l, e := a.fromQuery(ctx, qn, "app_id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAppDef: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with multiple matching AppDef
func (a *DBDeploymentLog) ByMultiAppDef(ctx context.Context, p []uint64) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByAppDef"
	l, e := a.fromQuery(ctx, qn, "app_id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAppDef: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeAppDef(ctx context.Context, p uint64) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeAppDef"
	l, e := a.fromQuery(ctx, qn, "app_id ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAppDef: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with matching BuildID
func (a *DBDeploymentLog) ByBuildID(ctx context.Context, p uint64) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByBuildID"
	l, e := a.fromQuery(ctx, qn, "buildid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with multiple matching BuildID
func (a *DBDeploymentLog) ByMultiBuildID(ctx context.Context, p []uint64) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByBuildID"
	l, e := a.fromQuery(ctx, qn, "buildid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeBuildID(ctx context.Context, p uint64) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeBuildID"
	l, e := a.fromQuery(ctx, qn, "buildid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with matching AutoDeployerHost
func (a *DBDeploymentLog) ByAutoDeployerHost(ctx context.Context, p string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByAutoDeployerHost"
	l, e := a.fromQuery(ctx, qn, "autodeployerhost = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAutoDeployerHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with multiple matching AutoDeployerHost
func (a *DBDeploymentLog) ByMultiAutoDeployerHost(ctx context.Context, p []string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByAutoDeployerHost"
	l, e := a.fromQuery(ctx, qn, "autodeployerhost in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAutoDeployerHost: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeAutoDeployerHost(ctx context.Context, p string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeAutoDeployerHost"
	l, e := a.fromQuery(ctx, qn, "autodeployerhost ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAutoDeployerHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with matching Started
func (a *DBDeploymentLog) ByStarted(ctx context.Context, p uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByStarted"
	l, e := a.fromQuery(ctx, qn, "started = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByStarted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with multiple matching Started
func (a *DBDeploymentLog) ByMultiStarted(ctx context.Context, p []uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByStarted"
	l, e := a.fromQuery(ctx, qn, "started in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByStarted: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeStarted(ctx context.Context, p uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeStarted"
	l, e := a.fromQuery(ctx, qn, "started ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByStarted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with matching Finished
func (a *DBDeploymentLog) ByFinished(ctx context.Context, p uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByFinished"
	l, e := a.fromQuery(ctx, qn, "finished = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByFinished: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with multiple matching Finished
func (a *DBDeploymentLog) ByMultiFinished(ctx context.Context, p []uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByFinished"
	l, e := a.fromQuery(ctx, qn, "finished in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByFinished: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeFinished(ctx context.Context, p uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeFinished"
	l, e := a.fromQuery(ctx, qn, "finished ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByFinished: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with matching Message
func (a *DBDeploymentLog) ByMessage(ctx context.Context, p string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByMessage"
	l, e := a.fromQuery(ctx, qn, "message = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with multiple matching Message
func (a *DBDeploymentLog) ByMultiMessage(ctx context.Context, p []string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByMessage"
	l, e := a.fromQuery(ctx, qn, "message in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMessage: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeMessage(ctx context.Context, p string) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeMessage"
	l, e := a.fromQuery(ctx, qn, "message ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with matching DeployAlgorithm
func (a *DBDeploymentLog) ByDeployAlgorithm(ctx context.Context, p uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByDeployAlgorithm"
	l, e := a.fromQuery(ctx, qn, "deployalgorithm = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeployAlgorithm: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBDeploymentLog" rows with multiple matching DeployAlgorithm
func (a *DBDeploymentLog) ByMultiDeployAlgorithm(ctx context.Context, p []uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByDeployAlgorithm"
	l, e := a.fromQuery(ctx, qn, "deployalgorithm in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeployAlgorithm: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBDeploymentLog) ByLikeDeployAlgorithm(ctx context.Context, p uint32) ([]*savepb.DeploymentLog, error) {
	qn := "DBDeploymentLog_ByLikeDeployAlgorithm"
	l, e := a.fromQuery(ctx, qn, "deployalgorithm ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeployAlgorithm: error scanning (%s)", e))
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
func (a *DBDeploymentLog) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.DeploymentLog, error) {
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

func (a *DBDeploymentLog) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.DeploymentLog, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBDeploymentLog) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.DeploymentLog, error) {
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
func (a *DBDeploymentLog) get_col_from_proto(p *savepb.DeploymentLog, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "r_binary" {
		return a.get_Binary(p)
	} else if colname == "app_id" {
		return a.get_AppDef_ID(p)
	} else if colname == "buildid" {
		return a.get_BuildID(p)
	} else if colname == "autodeployerhost" {
		return a.get_AutoDeployerHost(p)
	} else if colname == "started" {
		return a.get_Started(p)
	} else if colname == "finished" {
		return a.get_Finished(p)
	} else if colname == "message" {
		return a.get_Message(p)
	} else if colname == "deployalgorithm" {
		return a.get_DeployAlgorithm(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBDeploymentLog) Tablename() string {
	return a.SQLTablename
}

func (a *DBDeploymentLog) SelectCols() string {
	return "id,r_binary, app_id, buildid, autodeployerhost, started, finished, message, deployalgorithm"
}
func (a *DBDeploymentLog) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".r_binary, " + a.SQLTablename + ".app_id, " + a.SQLTablename + ".buildid, " + a.SQLTablename + ".autodeployerhost, " + a.SQLTablename + ".started, " + a.SQLTablename + ".finished, " + a.SQLTablename + ".message, " + a.SQLTablename + ".deployalgorithm"
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
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

