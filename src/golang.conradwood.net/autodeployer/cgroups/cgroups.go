package cgroups

import (
	"flag"
	"fmt"
	dm "golang.conradwood.net/apis/deploymonkey"
	dc "golang.conradwood.net/autodeployer/deployments"
	"golang.conradwood.net/go-easyops/utils"
	"os"
	"sync"
)

/*
Warning:
This will fail-late (on purpose). If config cgroup is called on a system with unsupported cgroup version
it will panic
*/
/*
cgroup v2:
mkdir /sys/fs/cgroup/[groupname]
echo +memory /sys/fs/cgroup/cgroups.subtree_controllers
mkdir /sys/fs/cgroup/[groupname]/tasks
echo [pid] /sys/fs/cgroup/[groupname]/cgroup.procs

*/

const (
	cgroupv1_memdir   = "/sys/fs/cgroup/memory"
	cgroupv2_testfile = "/sys/fs/cgroup/cgroup.controllers"
	cgroupv2_rootdir  = "/sys/fs/cgroup"
)

var (
	debug          = flag.Bool("debug_cgroups", false, "debug cgroups")
	cgroup_version = 0
	cgroupLock     sync.Mutex
)

type CGroupProc interface {
	Limits() *dm.Limits
	SetCgroup(int)  // store current cgroup number
	GetCgroup() int // retrieve current cgroup number
	GetPid() uint64
}

func init() {
	_, err := os.Stat(cgroupv1_memdir)
	if err == nil {
		cgroup_version = 1
		return
	}
	_, err = os.Stat(cgroupv2_testfile)
	if err == nil {
		cgroup_version = 2
		return
	}
}

func none(dc.Deployed) {
}
func ConfigCGroup(d CGroupProc) error {
	if cgroup_version == 1 {
		return ConfigCGroupV1(d)
	} else if cgroup_version == 2 {
		return ConfigCGroupV2(d)
	}
	panic(fmt.Sprintf("ConfigCgroup called on a system with an unsupported version of cgroup (%d)", cgroup_version))
}
func ConfigCGroupV2(d CGroupProc) error {
	cgroupLock.Lock()
	defer cgroupLock.Unlock()
	i := findFreeNum(cgroupv2_rootdir)
	cgdir := fmt.Sprintf("%s/%d", cgroupv2_rootdir, i)
	err := os.Mkdir(cgdir, 0770)
	if err != nil {
		return err
	}
	d.SetCgroup(i)

	//enable memory controller
	err = utils.WriteFile(fmt.Sprintf("%s/cgroup.subtree_control", cgdir), []byte("+memory\n"))
	if err != nil {
		RemoveCgroup(d)
		return err
	}

	// create task dir. we need one subtree for subtree setting to take effect
	taskdir := fmt.Sprintf("%s/tasks", cgdir)
	err = os.Mkdir(taskdir, 0770)
	if err != nil {
		RemoveCgroup(d)
		return err
	}

	maxmem := uint32(3000)
	if d.Limits() != nil && d.Limits().MaxMemory != 0 {
		maxmem = d.Limits().MaxMemory
	}

	// set max memory
	mem_filename := fmt.Sprintf("%s/memory.high", taskdir)
	err = utils.WriteFile(mem_filename, []byte(fmt.Sprintf("%dM\n", maxmem)))
	if err != nil {
		RemoveCgroup(d)
		return err
	}

	// assign pid to cgroup
	pids := fmt.Sprintf("%d\n", d.GetPid())
	task_filename := fmt.Sprintf("%s/cgroup.procs", taskdir)
	err = utils.WriteFile(task_filename, []byte(pids))
	if err != nil {
		RemoveCgroup(d)
		return err
	}
	return nil
}

func ConfigCGroupV1(d CGroupProc) error {
	cgroupLock.Lock()
	defer cgroupLock.Unlock()
	i := findFreeNum(cgroupv1_memdir)
	err := os.Mkdir(fmt.Sprintf("%s/%d", cgroupv1_memdir, i), 0770)
	if err != nil {
		return err
	}
	d.SetCgroup(i)

	maxmem := uint32(3000)
	if d.Limits() != nil && d.Limits().MaxMemory != 0 {
		maxmem = d.Limits().MaxMemory
	}
	mem_filename := fmt.Sprintf("%s/%d/memory.limit_in_bytes", cgroupv1_memdir, d.GetCgroup())
	err = utils.WriteFile(mem_filename, []byte(fmt.Sprintf("%dM\n", maxmem)))
	if err != nil {
		RemoveCgroup(d)
		return err
	}

	pids := fmt.Sprintf("%d\n", d.GetPid())
	task_filename := fmt.Sprintf("%s/%d/tasks", cgroupv1_memdir, d.GetCgroup())
	err = utils.WriteFile(task_filename, []byte(pids))
	if err != nil {
		RemoveCgroup(d)
		return err
	}
	procs_filename := fmt.Sprintf("%s/%d/cgroup.procs", cgroupv1_memdir, d.GetCgroup())
	err = utils.WriteFile(procs_filename, []byte(pids))
	if err != nil {
		RemoveCgroup(d)
		return err
	}
	return nil
}

// in a directory, find the lowest number that DOES NOT EXIST
func findFreeNum(dir string) int {
	i := 1
	for {
		fname := dir + fmt.Sprintf("/%d", i)
		if !utils.FileExists(fname) {
			return i
		}
		i++
	}
}
func RemoveCgroup(d CGroupProc) {
	if d.GetCgroup() != 0 {
		dir := fmt.Sprintf("%s/%d", cgroupv1_memdir, d.GetCgroup())
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Printf("unable to remove dir %s: %s\n", dir, err)
		} else {
			fmt.Printf("Cgroup %d removed\n", d.GetCgroup())
			d.SetCgroup(0)
		}
	}
}

func Debugf(txt string, args ...interface{}) {
	if !*debug {
		return
	}
	pid := os.Getpid()
	prefix := fmt.Sprintf("[cgroups pid=%d] ", pid)
	fmt.Printf(prefix+txt, args...)

}
func Errorf(txt string, args ...interface{}) {
	pid := os.Getpid()
	prefix := fmt.Sprintf("[cgroups pid=%d] ", pid)
	fmt.Printf(prefix+txt, args...)

}
