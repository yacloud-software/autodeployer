package deployer

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/autodeployer"
	"golang.conradwood.net/autodeployer/deployments"
)

func Deploy(ctx context.Context, du *deployments.Deployed, cr *pb.DeployRequest) (*pb.DeployResponse, error) {
	du.StartedWithContainer = true
	return nil, fmt.Errorf("Deploy with container not implemented yet")
}
