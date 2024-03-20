package deployplacements

import (
	"fmt"
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
	return fmt.Sprintf("%s on %s", dr.appdef.Binary, dr.sa.String())
}
func (dr *DeployRequest) AutodeployerHost() string {
	return dr.sa.Host()
}
func (dr *DeployRequest) AppDef() *pb.ApplicationDefinition {
	return dr.appdef
}
func Create_requests_for_app(group Group, app *pb.ApplicationDefinition, sas []*registry.ServiceAddress) ([]*DeployRequest, error) {
	var res []*DeployRequest
	ag := common.NewAutodeployerGroup(sas)
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
