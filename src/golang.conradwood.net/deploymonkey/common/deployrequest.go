package common

import (
	ad "golang.conradwood.net/apis/autodeployer"
	dm "golang.conradwood.net/apis/deploymonkey"
	"strings"
)

const (
	DEPLOYER_NAME = "deploymonkey"
)

func CreateInfoRequest() *ad.InfoRequest {
	res := &ad.InfoRequest{Deployer: DEPLOYER_NAME}
	return res
}

func CreateDeployRequest(group *dm.GroupDefinitionRequest, app *dm.ApplicationDefinition) *ad.DeployRequest {
	url := app.DownloadURL
	url = strings.ReplaceAll(url, "${BUILDID}", "latest")
	res := &ad.DeployRequest{
		Deployer:         DEPLOYER_NAME,
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
		res.Limits = &dm.Limits{}
	}
	if res.Limits.MaxMemory == 0 {
		res.Limits.MaxMemory = 3000
	}
	if group != nil {
		res.Groupname = group.GroupID
		res.Namespace = group.Namespace
	}
	return res
}
