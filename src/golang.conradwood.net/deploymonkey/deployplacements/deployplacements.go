package deployplacements

import (
	"fmt"
	"strings"

	ad "golang.conradwood.net/apis/autodeployer"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/apis/registry"
	"golang.conradwood.net/deploymonkey/common"
)

type DeployRequest struct {
	appdef *pb.ApplicationDefinition
	sa     *common.Deployer
}
type Group interface {
}

func (dr *DeployRequest) String() string {
	return fmt.Sprintf("%s,vers=%d,repo=%d,artefact=%d on %s", dr.appdef.Binary, dr.appdef.BuildID, dr.appdef.RepositoryID, dr.appdef.ArtefactID, dr.sa.String())
}
func (dr *DeployRequest) GetAutodeployerClient() ad.AutoDeployerClient {
	return dr.sa.GetClient()
}
func (dr *DeployRequest) AutodeployerHost() string {
	return dr.sa.Host()
}
func (dr *DeployRequest) Deployer() *common.Deployer {
	return dr.sa
}

// the url with variables resolved
func (dr *DeployRequest) DownloadURL() string {
	s := dr.AppDef().DownloadURL
	s = strings.ReplaceAll(s, "${BUILDID}", fmt.Sprintf("%d", dr.AppDef().BuildID))
	return s
}

// the url as defined in deploy.yaml
func (dr *DeployRequest) URL() string {
	return dr.AppDef().DownloadURL
}
func (dr *DeployRequest) AppDef() *pb.ApplicationDefinition {
	return dr.appdef
}
func Create_requests_for_app(group Group, app *pb.ApplicationDefinition, sas []*registry.ServiceAddress) ([]*DeployRequest, error) {
	var res []*DeployRequest
	ag, err := common.NewAutodeployerGroup(sas)
	if err != nil {
		return nil, err
	}
	if app.InstancesMeansPerAutodeployer {
		for _, s := range ag.FilterByMachine(app.Machines).Deployers() {
			for i := 0; i < int(app.Instances); i++ {
				dr := &DeployRequest{
					appdef: app,
					sa:     s,
				}
				res = append(res, dr)
			}
		}
		return res, nil
	}

	for len(res) < int(app.Instances) {
		if app.Machines == "" {
			app.Machines = "worker"
		}
		for _, s := range ag.FilterByMachine(app.Machines).Deployers() {
			dr := &DeployRequest{
				appdef: app,
				sa:     s,
			}
			res = append(res, dr)
			if len(res) >= int(app.Instances) {
				break
			}

		}
	}
	return res, nil

}
