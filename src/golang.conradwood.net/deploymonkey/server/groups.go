package main

import (
	"context"
	"database/sql"
	//	"errors"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/db"
)

type DBGroup interface {
	//	GetDeployedVersion() uint32
	//	GetPendingVersion() uint32
	//
	// SetApplications(x []*pb.ApplicationDefinition)
	// GetGroupDef() *pb.GroupDefinitionRequest
}
type XDBGroup struct {
	id              int
	DeployedVersion int
	PendingVersion  int
	groupDef        *pb.GroupDefinitionRequest
}

func (db *XDBGroup) GetGroupDef() *pb.GroupDefinitionRequest {
	return db.groupDef
}

func (db *XDBGroup) GetDeployedVersion() uint32 {
	return uint32(db.DeployedVersion)
}
func (db *XDBGroup) GetPendingVersion() uint32 {
	return uint32(db.PendingVersion)
}

func XgetGroupForAppByID(ctx context.Context, appid int) (DBGroup, error) {
	rows, err := dbcon.QueryContext(ctx, "getgroupforapp", "select appgroup.id,appgroup.groupname,appgroup.deployedversion,appgroup.pendingversion,appgroup.namespace from appgroup,group_version,lnk_app_grp where appgroup.id = group_version.group_id and lnk_app_grp.group_version_id = group_version.id and lnk_app_grp.app_id=$1", appid)
	if err != nil {
		fmt.Printf("Failed to get for app #%d: %s\n", appid, err)
		return nil, err
	}
	return XgetGroupFromDatabaseByRow(rows)

}

// get the group with given id from database. if no such group will return nil
func XgetGroupFromDatabaseByID(ctx context.Context, id int) (DBGroup, error) {
	rows, err := dbcon.QueryContext(ctx, "getgroupbyid", "SELECT id,groupname,deployedversion,pendingversion,namespace from appgroup where id=$1", id)
	if err != nil {
		fmt.Printf("Failed to get group #%d: %s\n", id, err)
		return nil, err
	}
	return XgetGroupFromDatabaseByRow(rows)
}

// get the group with given name from database. if no such group will return nil
func XgetGroupFromDatabase(ctx context.Context, nameSpace string, groupName string) (DBGroup, error) {
	rows, err := dbcon.QueryContext(ctx, "getgroup", "SELECT id,groupname,deployedversion,pendingversion,namespace from appgroup where groupname=$1 and namespace=$2", groupName, nameSpace)
	if err != nil {
		fmt.Printf("Failed to get groupname %s\n", groupName)
		return nil, err
	}
	return XgetGroupFromDatabaseByRow(rows)
}
func XgetGroupFromDatabaseByRow(rows *sql.Rows) (DBGroup, error) {
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	gdr := &pb.GroupDefinitionRequest{}
	d := XDBGroup{groupDef: gdr}
	err := rows.Scan(&d.id, &gdr.GroupID, &d.DeployedVersion, &d.PendingVersion, &gdr.Namespace)
	if err != nil {
		fmt.Printf("Failed to get row for group: %s\n", err)
		return nil, err
	}

	return &d, nil

}
func XcreateGroup(ctx context.Context, nameSpace string, groupName string) (DBGroup, error) {
	ad := &pb.AppGroup{Groupname: groupName, Namespace: nameSpace}
	_, err := db.DefaultDBAppGroup().Save(context.Background(), ad)
	if err != nil {
		return nil, err
	}
	return XgetGroupFromDatabase(ctx, nameSpace, groupName)

}

// create a new group version, return versionID
func createGroupVersion(ctx context.Context, nameSpace string, groupName string, def []*pb.ApplicationDefinition) (string, error) {
	panic("no group")
}
