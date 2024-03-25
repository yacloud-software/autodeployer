package main

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	//	dc "golang.conradwood.net/deploymonkey/common"
	"golang.conradwood.net/go-easyops/errors"
	// "strconv"
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
		return nil, fmt.Errorf("database not open")
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
	return nil, errors.NotImplemented(ctx, "UpdateApp")
}

// updates all apps in a repo to a new buildid
func (s *DeployMonkey) UpdateRepo(ctx context.Context, cr *pb.UpdateRepoRequest) (*pb.GroupDefResponse, error) {
	return nil, errors.NotImplemented(ctx, "UpdateRepo")
}
