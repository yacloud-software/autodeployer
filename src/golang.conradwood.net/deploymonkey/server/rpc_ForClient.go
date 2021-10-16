package main

import (
	"context"
	"errors"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	dc "golang.conradwood.net/deploymonkey/common"
	"strconv"
)

// stuff that's used by deploymonkey (only)

// used by deploymonkey-client
func (s *DeployMonkey) ListVersionsByName(ctx context.Context, cr *pb.ListVersionByNameRequest) (*pb.GetAppVersionsResponse, error) {
	n := cr.Name
	fmt.Printf("Getting apps for repo \"%s\"", n)
	q := fmt.Sprintf(`
SELECT lnk_app_grp.app_id,lnk_app_grp.group_version_id,group_version.created from lnk_app_grp 
inner join applicationdefinition on applicationdefinition.id = lnk_app_grp.app_id 
inner join group_version on group_version.id = lnk_app_grp.group_version_id
inner join appgroup on appgroup.id = group_version.group_id
where applicationdefinition.r_binary like '%%%s%%'
OR appgroup.namespace  like '%%%s%%'
OR appgroup.groupname  like '%%%s%%'
order by group_version_id desc
`, n, n, n)
	avds, err := listVersions(ctx, q)
	if err != nil {
		return nil, err
	}
	res := pb.GetAppVersionsResponse{}
	for _, avd := range avds {
		gar := pb.GetAppResponse{
			Created:     avd.gv.Created.Unix(),
			VersionID:   int64(avd.gv.Version),
			Application: avd.appdef,
		}
		res.Apps = append(res.Apps, &gar)
	}
	return &res, nil
}
func (s *DeployMonkey) ListVersionsForGroup(ctx context.Context, cr *pb.ListVersionRequest) (*pb.GetAppVersionsResponse, error) {
	fmt.Printf("Getting apps for %d\n", cr.RepositoryID)
	q := fmt.Sprintf(`
SELECT lnk_app_grp.app_id,lnk_app_grp.group_version_id,group_version.created from lnk_app_grp 
inner join applicationdefinition on applicationdefinition.id = lnk_app_grp.app_id 
inner join group_version on group_version.id = lnk_app_grp.group_version_id
where applicationdefinition.repositoryid = %d 
order by group_version_id desc
`, cr.RepositoryID)
	avds, err := listVersions(ctx, q)
	if err != nil {
		return nil, err
	}
	res := pb.GetAppVersionsResponse{}
	for _, avd := range avds {
		gar := pb.GetAppResponse{
			Created:     avd.gv.Created.Unix(),
			VersionID:   int64(avd.gv.Version),
			Application: avd.appdef,
		}
		res.Apps = append(res.Apps, &gar)
	}
	return &res, nil
}

// the query must return (in that order) applicationdefinition.id, lnk_app_grp.group_version_id, group_version,created
func listVersions(ctx context.Context, q string) ([]*appVersionDef, error) {
	if dbcon == nil {
		return nil, errors.New("database not open")
	}
	fmt.Printf("Getting versions\n")
	// this query gives us the version in lnk_app_grp.group_version_id
	rows, err := dbcon.QueryContext(TEMPCONTEXT(), "list_versions", q+` limit $1`, *limit)
	if err != nil {
		fmt.Printf("Failed to query app: %s\n", err)
		return nil, err
	}
	defer rows.Close()
	var res []*appVersionDef
	for rows.Next() {
		var app_id uint64
		gv := groupVersion{}
		err = rows.Scan(&app_id, &gv.Version, &gv.Created)
		if err != nil {
			fmt.Printf("1. Failed to get apps for repo (rows):%s\n", err)
			return nil, err
		}
		ad, err := loadApp(ctx, app_id)
		if err != nil {
			fmt.Printf("2. Failed to get apps for repo (rows):%s\n", err)
			return nil, err
		}
		fmt.Printf("Version #%d: AppID: %d, BuildID: %d, repository=%d (%v)\n", gv.Version, ad.ID, ad.BuildID, ad.RepositoryID, gv.Created)
		r := appVersionDef{
			appdef: ad,
			gv:     &gv,
		}
		res = append(res, &r)
	}
	return res, nil
}

// updates a single app to a new version
func (s *DeployMonkey) UpdateApp(ctx context.Context, cr *pb.UpdateAppRequest) (*pb.GroupDefResponse, error) {
	if cr.App.BuildID == 0 {
		return nil, errors.New("BuildID 0 is invalid")
	}
	if cr.Namespace == "" {
		return nil, errors.New("Namespace required")
	}
	if cr.GroupID == "" {
		return nil, errors.New("GroupID required")
	}
	if cr.App.RepositoryID == 0 {
		return nil, errors.New("App Repository required")
	}
	fmt.Printf("Request to update app:\n")
	dc.PrintApp(cr.App)

	cur, err := getGroupFromDatabase(ctx, cr.Namespace, cr.GroupID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get group from db: %s", err))
	}
	if cur == nil {
		return nil, errors.New(fmt.Sprintf("No such group: (%s,%s)", cr.Namespace, cr.GroupID))
	}

	lastVersion, err := getGroupLatestVersion(ctx, cr.Namespace, cr.GroupID)
	if err != nil {
		return nil, err
	}
	apps, err := loadAppGroupVersion(ctx, lastVersion)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get apps for version %d from db: %s", cur.DeployedVersion, err))
	}
	fmt.Printf("Loaded Group from database: \n")
	cur.groupDef.Applications = apps
	dc.PrintGroup(cur.groupDef)
	// now find the app we want to update:
	foundone := false
	for _, app := range apps {
		if isSame(app, cr.App) {
			fmt.Printf("Updating app: %d\n", app.RepositoryID)
			m := mergeApp(app, cr.App)
			if !m {
				fmt.Printf("Nothing to update, app is already up to date\n")
				r := pb.GroupDefResponse{Result: pb.GroupResponseStatus_NOCHANGE}
				return &r, nil
			}
			foundone = true
			break
		}
	}
	if !foundone {
		return nil, errors.New(fmt.Sprintf("There is no app \"%d\" in group (%s,%s)", cr.App.RepositoryID, cr.Namespace, cr.GroupID))
	}
	cur.groupDef.Applications = apps
	fmt.Printf("Updated Group: \n")
	cur.groupDef.Applications = apps
	dc.PrintGroup(cur.groupDef)

	sv, err := createGroupVersion(ctx, cr.Namespace, cr.GroupID, apps)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to create a new group version: %s", err))
	}
	fmt.Printf("Created group version: %s\n", sv)
	version, err := strconv.Atoi(sv)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("version group not a number (%s)? BUG!: %s", sv, err))
	}

	applyVersion(version)
	if err != nil {
		return nil, err
	}
	updateDeployedVersionNumber(version)
	r := pb.GroupDefResponse{Result: pb.GroupResponseStatus_CHANGEACCEPTED}
	return &r, nil
}

// updates all apps in a repo to a new buildid
func (s *DeployMonkey) UpdateRepo(ctx context.Context, cr *pb.UpdateRepoRequest) (*pb.GroupDefResponse, error) {
	if cr.Namespace == "" {
		return nil, errors.New("Namespace required")
	}
	if cr.GroupID == "" {
		return nil, errors.New("GroupID required")
	}
	if cr.RepositoryID == 0 {
		return nil, errors.New("App Repository required")
	}
	fmt.Printf("Updating all apps in repository %d in (%s,%s) to buildid: %d\n", cr.RepositoryID,
		cr.Namespace, cr.GroupID, cr.BuildID)
	cur, err := getGroupFromDatabase(ctx, cr.Namespace, cr.GroupID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get group from db: %s", err))
	}
	if cur == nil {
		return nil, errors.New(fmt.Sprintf("No such group: (%s,%s)", cr.Namespace, cr.GroupID))
	}

	lastVersion, err := getGroupLatestVersion(ctx, cr.Namespace, cr.GroupID)
	if err != nil {
		return nil, err
	}
	apps, err := loadAppGroupVersion(ctx, lastVersion)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get apps for version %d from db: %s", cur.DeployedVersion, err))
	}
	fmt.Printf("Loaded Group from database: \n")
	cur.groupDef.Applications = apps
	dc.PrintGroup(cur.groupDef)
	// now find the app we want to update:
	foundone := false
	for _, app := range apps {
		if app.RepositoryID != cr.RepositoryID {
			continue
		}
		fmt.Printf("Updating app: %d\n", app.RepositoryID)
		app.BuildID = cr.BuildID
		foundone = true
	}
	if !foundone {
		fmt.Printf("Nothing to update, app is already up to date\n")
		r := pb.GroupDefResponse{Result: pb.GroupResponseStatus_NOCHANGE}
		return &r, nil
	}
	cur.groupDef.Applications = apps
	fmt.Printf("Updated Group: \n")
	cur.groupDef.Applications = apps
	dc.PrintGroup(cur.groupDef)

	sv, err := createGroupVersion(ctx, cr.Namespace, cr.GroupID, apps)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to create a new group version: %s", err))
	}
	fmt.Printf("Created group version: %s\n", sv)
	version, err := strconv.Atoi(sv)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("version group not a number (%s)? BUG!: %s", sv, err))
	}
	applyVersion(version)
	if err != nil {
		return nil, err
	}
	updateDeployedVersionNumber(version)
	r := pb.GroupDefResponse{Result: pb.GroupResponseStatus_CHANGEACCEPTED}
	return &r, nil

}
