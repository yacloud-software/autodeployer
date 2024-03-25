package main

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/db"
	"sync"
	"time"
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

var (
	grouplock    sync.Mutex
	groupHandler *Group2Handler
)

type Group2Handler struct {
	db_gv  *db.DBGroupVersion
	db_ag  *db.DBAppGroup
	db_lag *db.DBLinkAppGroup
}
type creator interface {
	CreateTable(ctx context.Context) error
}

func start_group2_handler() error {
	g2 := &Group2Handler{
		db_gv:  db.NewDBGroupVersion(dbcon),
		db_ag:  db.NewDBAppGroup(dbcon),
		db_lag: db.NewDBLinkAppGroup(dbcon),
	}
	g2.db_gv.SQLTablename = "group_version"
	g2.db_ag.SQLTablename = "appgroup"
	g2.db_lag.SQLTablename = "lnk_app_grp"
	var creates []creator
	creates = append(creates, g2.db_gv)
	creates = append(creates, g2.db_ag)
	creates = append(creates, g2.db_lag)
	fmt.Printf("Creating group tables:\n")
	for _, c := range creates {
		ctx := context.Background()
		err := c.CreateTable(ctx)
		if err != nil {
			return err
		}
	}
	groupHandler = g2
	return nil
}
func (g *Group2Handler) GroupByID(ctx context.Context, ID uint64) (*pb.AppGroup, error) {
	ags, err := g.db_ag.ByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	return ags, nil
}
func (g *Group2Handler) CreateGroupVersion(ctx context.Context, group *pb.GroupDefinitionRequest) (*pb.GroupVersion, error) {
	appgroup, err := g.FindOrCreateAppGroupByNamespace(ctx, group.Namespace)
	if err != nil {
		return nil, err
	}
	fmt.Printf("AppGroup: %d\n", appgroup.ID)
	// create new version of this group
	gv := &pb.GroupVersion{
		GroupID:          appgroup,
		CreatedTimestamp: uint32(time.Now().Unix()),
	}
	_, err = g.db_gv.Save(ctx, gv)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Groupversion: %d\n", gv.ID)
	// add the applications to the group version
	for _, app := range group.Applications {
		id, err := db.DefaultDBApplicationDefinition().Save(ctx, app)
		if err != nil {
			return nil, err
		}
		fmt.Printf("New application #%d\n", id)
		lag := &pb.LinkAppGroup{
			GroupVersion: gv,
			App:          app,
		}
		_, err = g.db_lag.Save(ctx, lag)
		if err != nil {
			return nil, err
		}
	}
	return gv, nil
}
func (g *Group2Handler) FindAppGroupByNamespace(ctx context.Context, namespace string) (*pb.AppGroup, error) {
	ags, err := g.db_ag.ByNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}
	if len(ags) == 0 {
		return nil, nil
	}
	return ags[0], nil

}
func (g *Group2Handler) FindOrCreateAppGroupByNamespace(ctx context.Context, namespace string) (*pb.AppGroup, error) {
	grouplock.Lock()
	defer grouplock.Unlock()
	ags, err := g.db_ag.ByNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}
	if len(ags) != 0 {
		return ags[0], nil
	}
	appgroup := &pb.AppGroup{
		Namespace: namespace,
		Groupname: "testing",
	}
	_, err = g.db_ag.Save(ctx, appgroup)
	if err != nil {
		return nil, err
	}
	return appgroup, nil
}

func (g *Group2Handler) GetGroupForApp(ctx context.Context, app *pb.ApplicationDefinition) (*pb.AppGroup, error) {
	lgroups, err := g.db_lag.ByApp(ctx, app.ID)
	if err != nil {
		return nil, err
	}
	if len(lgroups) == 0 {
		return nil, fmt.Errorf("no group for app %d", app.ID)
	}
	lgr := lgroups[0]

	gv, err := g.db_gv.ByID(ctx, lgr.GroupVersion.ID)
	if err != nil {
		return nil, err
	}

	group, err := g.db_ag.ByID(ctx, gv.GroupID.ID)
	if err != nil {
		return nil, err
	}
	return group, nil
}
