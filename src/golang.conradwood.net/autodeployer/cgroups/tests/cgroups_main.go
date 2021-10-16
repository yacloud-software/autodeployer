package main

// this is not a go test, because we have to run it as root, which
// makes it annoying to use go test

import (
	"context"
	"flag"
	"fmt"
	dm "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/autodeployer/cgroups"
	//	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/utils"
	"os"
	"os/exec"
	"syscall"
)

type tproc struct {
	cg  int
	pid uint64
}

func (t *tproc) Limits() *dm.Limits {
	return nil
}
func (t *tproc) SetCgroup(i int) {
	t.cg = i
}
func (t *tproc) GetCgroup() int {
	return t.cg
}
func (t *tproc) GetPid() uint64 {
	return t.pid
}
func main() {
	flag.Parse()
	//	fmt.Printf("testing proc\n")
	i := cgroups.Startup()
	fmt.Printf("Startup returned %d\n", i)
	if i == 0 {
		return
	}
	du := &tproc{}
	du.pid = uint64(os.Getpid())
	err := cgroups.ConfigCGroup(du)
	utils.Bail("failed to set cgroup: %s", err)

	ctx := context.Background()
	c := exec.CommandContext(ctx, "/bin/bash", "--norc", "-i")
	c.Env = []string{
		"foo=bar",
		"PS1=m1shell> ",
		"PS2=m2shell> ",
	}
	c.Stdin = os.Stdin
	c.Stdout = &cgroups.PrintingWriter{}
	c.Stderr = &cgroups.PrintingWriter{}
	c.SysProcAttr = &syscall.SysProcAttr{Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNET}
	err = c.Run()
	utils.Bail("error executing command", err)

	/*
		l := linux.New()
		l.AddNamespace("NEWUSER")
		//	out, err := l.SafelyExecuteWithDir([]string{"ls", "-l", "/tmp"}, "/", nil)
		out, err := l.SafelyExecuteWithDir([]string{"/bin/sh"}, "/", os.Stdin)
		fmt.Printf("output: %s\n", out)
		utils.Bail("error executing command", err)
	*/
}
