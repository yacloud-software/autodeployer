package common

import (
	"flag"
	ad "golang.conradwood.net/apis/autodeployer"
	dm "golang.conradwood.net/apis/deploymonkey"
	"strings"
)

const (
	DEFAULT_PRIORITY = 5
)

var (
	deployer_name = flag.String("deployer_name", "deploymonkey", "name of a deploymonkey, can be used to partition autodeployers between deploymonkeys")
)

func CreateInfoRequest() *ad.InfoRequest {
	res := &ad.InfoRequest{Deployer: *deployer_name}
	return res
}

func CreateDeployRequest(group *dm.GroupDefinitionRequest, app *dm.ApplicationDefinition) *ad.DeployRequest {
	app.Limits = AppLimits(app) // if non assigned in deploy.yaml, create a default applimits, otherwise use deploy.yaml values
	url := app.DownloadURL
	url = strings.ReplaceAll(url, "${BUILDID}", "latest")
	res := &ad.DeployRequest{
		Deployer:         *deployer_name,
		DownloadURL:      url,
		DownloadUser:     app.DownloadUser,
		DownloadPassword: app.DownloadPassword,
		Binary:           app.Binary,
		Args:             app.Args,
		RepositoryID:     app.RepositoryID,
		BuildID:          app.BuildID,
		DeploymentID:     app.DeploymentID,
		DeployType:       app.DeployType,
		StaticTargetDir:  app.StaticTargetDir,
		Public:           app.Public,
		AutoRegistration: app.AutoRegs,
		Limits:           app.Limits,
		AppReference:     &dm.AppReference{AppDef: app},
		ArtefactID:       app.ArtefactID,
	}
	if res.Limits == nil {
		panic("No limits assigned, despite running Applimits() earlier on\n")
	}

	if group != nil {
		res.Groupname = group.GroupID
		res.Namespace = group.Namespace
	}
	return res
}
