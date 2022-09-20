package packages

import (
	"context"
	"flag"
	pb "golang.conradwood.net/apis/autodeployer"
)

var (
	debug    = flag.Bool("debug_packages", false, "debug package manager")
	noauth   = flag.Bool("packages_allow_all", false, "if true allow everyone to install packages")
	use_sudo = flag.Bool("packages_use_sudo", false, "if true use sudo to access OS package manager")
)

type PackageManager interface {
	CheckPackage(ctx context.Context, req *pb.PackageInstallRequest) (*pb.PackageInstallResponse, error)
	InstallPackage(ctx context.Context, req *pb.PackageInstallRequest) (*pb.PackageInstallResponse, error)
}

func New() PackageManager {
	return &deb{}
}
