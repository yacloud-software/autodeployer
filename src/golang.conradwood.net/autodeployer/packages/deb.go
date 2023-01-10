package packages

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/autodeployer"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/linux"
	"sync"
	"time"
)

var (
	deblock sync.Mutex
)

type deb struct {
}

func (s *deb) CheckPackage(ctx context.Context, req *pb.PackageInstallRequest) (*pb.PackageInstallResponse, error) {
	if !*noauth {
		err := errors.NeedsRoot(ctx)
		if err != nil {
			return nil, err
		}
	}
	l := linux.NewWithContext(ctx)
	com := []string{"dpkg", "-s", req.Name}
	out, err := l.SafelyExecuteWithDir(com, "/tmp", nil)
	if *debug {
		fmt.Printf("Output: %s\n", out)
	}
	if err != nil {
		return &pb.PackageInstallResponse{
			Name:      req.Name,
			Installed: false,
		}, nil
	}
	return &pb.PackageInstallResponse{
		Name:      req.Name,
		Installed: true,
	}, nil

}
func (s *deb) InstallPackage(ctx context.Context, req *pb.PackageInstallRequest) (*pb.PackageInstallResponse, error) {
	if !*noauth {
		err := errors.NeedsRoot(ctx)
		if err != nil {
			return nil, err
		}
	}
	res, err := s.CheckPackage(ctx, req)
	if err != nil {
		return nil, err
	}
	go install_debian_package(req.Name)
	return res, nil

}
func install_debian_package(name string) {
	deblock.Lock()
	defer deblock.Unlock()
	l := linux.New()
	l.SetMaxRuntime(time.Duration(600) * time.Second)
	com := add_sudo([]string{"apt-get", "update"})
	out, err := l.SafelyExecuteWithDir(com, "/tmp", nil)
	if *debug {
		fmt.Printf("Output: %s\n", out)
	}
	if err != nil {
		fmt.Printf("apt-get install -y %s failed: %s\n", name, err)
		return
	}
	l = linux.New()
	l.SetMaxRuntime(time.Duration(600) * time.Second)
	com = add_sudo([]string{"apt-get", "install", "-y", name})
	out, err = l.SafelyExecuteWithDir(com, "/tmp", nil)
	if *debug {
		fmt.Printf("Output: %s\n", out)
	}
	if err != nil {
		fmt.Printf("apt-get install -y %s failed: %s\n", name, err)
	}
}

func add_sudo(sx []string) []string {
	if !*use_sudo {
		return sx
	}
	return append([]string{"sudo"}, sx...)
}
