package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/db"
)

/*
   Apps are always in a group. a group has 0 or more apps. This relationship is defined in the deploy.yaml and reflected in the database.
   Groups are versioned - a new version of an appdef creates a new groupVersion.
   a new app version also creates a new applicationdefinition row.
   the new app version is linked to a group version with lnk_app_grp.
   lnk_app_grp.app_id           -> applicationdefinition.id
   lnk_app_grp.group_version_id -> group_version.id
   group_version.group_id       -> group.id
*/

type DBGroup interface {
	GetDeployedVersion() uint32
	GetPendingVersion() uint32
}
type XDBGroup struct {
	id              int
	DeployedVersion int
	PendingVersion  int
	groupDef        *pb.GroupDefinitionRequest
}

func getGroupForAppByID(ctx context.Context, appid int) (DBGroup, error) {
	rows, err := dbcon.QueryContext(ctx, "getgroupforapp", "select appgroup.id,appgroup.groupname,appgroup.deployedversion,appgroup.pendingversion,appgroup.namespace from appgroup,group_version,lnk_app_grp where appgroup.id = group_version.group_id and lnk_app_grp.group_version_id = group_version.id and lnk_app_grp.app_id=$1", appid)
	if err != nil {
		fmt.Printf("Failed to get for app #%d: %s\n", appid, err)
		return nil, err
	}
	return getGroupFromDatabaseByRow(rows)

}

// get the group with given id from database. if no such group will return nil
func getGroupFromDatabaseByID(ctx context.Context, id int) (DBGroup, error) {
	rows, err := dbcon.QueryContext(ctx, "getgroupbyid", "SELECT id,groupname,deployedversion,pendingversion,namespace from appgroup where id=$1", id)
	if err != nil {
		fmt.Printf("Failed to get group #%d: %s\n", id, err)
		return nil, err
	}
	return getGroupFromDatabaseByRow(rows)
}

// get the group with given name from database. if no such group will return nil
func getGroupFromDatabase(ctx context.Context, nameSpace string, groupName string) (DBGroup, error) {
	rows, err := dbcon.QueryContext(ctx, "getgroup", "SELECT id,groupname,deployedversion,pendingversion,namespace from appgroup where groupname=$1 and namespace=$2", groupName, nameSpace)
	if err != nil {
		fmt.Printf("Failed to get groupname %s\n", groupName)
		return nil, err
	}
	return getGroupFromDatabaseByRow(rows)
}
func getGroupFromDatabaseByRow(rows *sql.Rows) (DBGroup, error) {
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
func createGroup(ctx context.Context, nameSpace string, groupName string) (DBGroup, error) {
	ad := &pb.AppGroup{Groupname: groupName, Namespace: nameSpace}
	_, err := db.DefaultDBAppGroup().Save(context.Background(), ad)
	if err != nil {
		return nil, err
	}
	return getGroupFromDatabase(ctx, nameSpace, groupName)

}

// create a new group version, return versionID
func createGroupVersion(ctx context.Context, nameSpace string, groupName string, def []*pb.ApplicationDefinition) (string, error) {
	var id int
	r, err := getGroupFromDatabase(ctx, nameSpace, groupName)
	if err != nil {
		return "", fmt.Errorf("[1]createGroupVersion():%s", err)
	}
	if r.groupDef.GroupID == "" {
		// had no row!
		r, err = createGroup(ctx, nameSpace, groupName)
		if err != nil {
			return "", fmt.Errorf("[2]createGroupVersion():%s", err)
		}
	}
	//	gv := &pb.GroupVersion{GroupID: r.Proto()}
	err = dbcon.QueryRowContext(TEMPCONTEXT(), "newgroupversion", "INSERT into group_version (group_id) values ($1) RETURNING id", r.id).Scan(&id)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to insert group_version: %s", err))
	}
	versionId := id
	fmt.Printf("New Version: %d for Group #%d\n", versionId, r.id)
	for _, ad := range def {
		fmt.Printf("Saving: %v (alwayson=%v,critical=%v)\n", ad, ad.AlwaysOn, ad.Critical)
		id, err := saveApp(ad)
		if err != nil {
			return "", err
		}
		fmt.Printf("Inserted App #%s\n", id)
		_, err = dbcon.ExecContext(TEMPCONTEXT(), "lnkappgrp", "INSERT into lnk_app_grp (group_version_id,app_id) values ($1,$2)", versionId, id)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Failed to add application to new version: %s", err))
		}
	}
	return fmt.Sprintf("%d", versionId), nil
}
