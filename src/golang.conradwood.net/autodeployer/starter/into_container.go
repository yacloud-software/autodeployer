package starter

import (
	"fmt"
	//	dm "golang.conradwood.net/apis/deploymonkey"
	ad "golang.conradwood.net/apis/autodeployer"
	//	"golang.conradwood.net/go-easyops/utils"
	"os"
	"syscall"
)

func intoContainer(sp *ad.StartupResponse) error {
	ad := sp.AppReference.AppDef
	ct := ad.Container
	if ct == nil {
		return nil
	}
	path := sp.WorkingDir
	os.Chdir(path)
	fmt.Printf("Container working directory: \"%s\"\n", path)
	err := syscall.Chroot(path)
	if err != nil {
		return fmt.Errorf("Chroot failed (%w)", err)
	}
	fmt.Printf("Container chrooted to %s\n", path)

	return nil
}
