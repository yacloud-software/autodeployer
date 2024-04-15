package common

import (
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	pb "golang.conradwood.net/apis/deploymonkey"
)

type ActualDeployment struct {
	deployer *Deployer
	app      *ad.DeployedApp
}

/*
quirky best-effort algorithm to find deployed apps that "match" an application definition, irrespective of version
*/
func FindByAppDef(appdef *pb.ApplicationDefinition) []*ActualDeployment {
	deplock.Lock()
	defer deplock.Unlock()
	var apps []*ActualDeployment
	fmt.Printf("Finding other deployments for app repo=%d, artefact=%d, binary=%s, deploymentid=%s\n", appdef.RepositoryID, appdef.ArtefactID, appdef.Binary, appdef.DeploymentID)
	if appdef.DeploymentID == "" {
		panic("findbyappdef with no deploymentid is not supported")
	}
	for _, d := range deployers {
		if d.lastInfoResponse == nil {
			continue
		}
		for _, deployedapp := range d.lastInfoResponse.Apps {
			dr := deployedapp.DeployRequest
			if dr == nil {
				continue
			}
			appref := dr.AppReference
			if appref == nil {
				continue
			}
			appd := appref.AppDef
			if appd == nil {
				continue
			}
			//fmt.Printf("   comparing with app repo=%d, artefact=%d, binary=%s, deploymentid=%s\n", appd.RepositoryID, appd.ArtefactID, appd.Binary, appd.DeploymentID)
			// we got an applicationdefinition from autodeployer and deploymonkey - compare them
			if appd.RepositoryID != appdef.RepositoryID {
				continue
			}
			if appd.ArtefactID != appdef.ArtefactID {
				continue
			}
			if appd.Binary != appdef.Binary {
				continue
			}
			if appd.DeploymentID != appdef.DeploymentID {
				continue
			}
			ade := &ActualDeployment{app: deployedapp, deployer: d}
			apps = append(apps, ade)
		}
	}
	return apps
}

func (ade *ActualDeployment) Deployer() *Deployer {
	return ade.deployer
}
func (ade *ActualDeployment) DeployedApp() *ad.DeployedApp {
	return ade.app
}
func (ade *ActualDeployment) AppDef() *pb.ApplicationDefinition {
	dr := ade.app.DeployRequest
	if dr == nil {
		return nil
	}
	appref := dr.AppReference
	if appref == nil {
		return nil
	}
	appd := appref.AppDef
	if appd == nil {
		return nil
	}
	return appd

}
