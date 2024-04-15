package deployq

import (
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	"golang.conradwood.net/deploymonkey/common"
	"golang.conradwood.net/go-easyops/authremote"
)

func stop_app(deployer *common.Deployer, id string) error {
	depl := deployer
	appid := id
	ctx := authremote.Context()
	ur := &ad.UndeployRequest{Block: true, ID: appid}
	_, err := depl.GetClient().Undeploy(ctx, ur)
	if err == nil {
		fmt.Printf("Undeployed app %s on autodeployer %s\n", appid, depl.Host())
	} else {
		fmt.Printf("Failed to undeploy app %s on autodeployer %s: %s\n", appid, depl.Host(), err)
	}
	return err
}
