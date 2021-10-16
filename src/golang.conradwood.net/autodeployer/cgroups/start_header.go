package cgroups

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const (
	STARTUP_MARKER = `muam4Iuhuphohphi4cod4phai9xe8leu8Thai3Mu7aej9aequ4Or2AiK4coongie6`
)

type StartupState struct {
	Forks int // how often did we fork() ?
}

// this function must be called at startup (to set pivot_root and mounts and stuff before handing
// control to the actual process
// the function takes a single startup argument, encoded
// it will fork() and execute this program (by using os.Arg[0]).
// the original code will block until the child has completed and then return 0
// the forked code will return with != 0
func Startup() int {
	fmt.Printf("this code is unsafe. suspected of deleting '/' on the computer it runs on. exiting now\n")
	os.Exit(10)
	Debugf("Startup code\n")
	state := &StartupState{}
	for _, a := range os.Args {
		//Debugf("Arg %d: %s\n", i, a)
		if strings.Contains(a, STARTUP_MARKER) {
			deser(a[len(STARTUP_MARKER):], state)
		}
	}
	Debugf("State: forks=%d\n", state.Forks)

	if state.Forks > 2 {
		err := do_stuff_after_fork()
		if err != nil {
			Errorf("failed to do stuff after fork: %s\n", err)
		}
		return 1
	}
	myprg := os.Args[0]
	Debugf("Forking self (%s)\n", myprg)
	state.Forks++
	arg, err := ser(state)
	if err != nil {
		Errorf("failed to serialise state: %s\n", err)
		return 1
	}

	// first fork?
	if state.Forks == 1 {
		// pivot_root fails with a shared mounted root filesystem
		err = syscall.Unshare(syscall.CLONE_NEWNS)
		if err != nil {
			Errorf("failed to unshare: %s\n", err)
			return 0
		}
		err = syscall.Mount("none", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, "")
		if err != nil {
			Errorf("failed to remount root", err)
		}

	}

	var newargs []string
	// strip previous startup markers
	for _, a := range os.Args {
		if strings.Contains(a, STARTUP_MARKER) {
			continue
		}
		newargs = append(newargs, a)
	}
	newargs = append(newargs[1:], STARTUP_MARKER+arg) // without binary name

	cmd := exec.Command(myprg, newargs...)

	if state.Forks > 1 {
		Debugf("Setting clone_newns\n")
		cmd.SysProcAttr = &syscall.SysProcAttr{Cloneflags: syscall.CLONE_NEWNS}
	}
	//	cmd.SysProcAttr = &syscall.SysProcAttr{Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNET}
	cmd.Stdin = os.Stdin
	cmd.Stdout = &PrintingWriter{}
	cmd.Stderr = &PrintingWriter{}
	err = cmd.Run()
	if err != nil {
		Errorf("Failed to run command (%s)\n", err)
	} else {
		Debugf("Re-exec completed\n")
	}
	return 0
}

type PrintingWriter struct {
}

func (p *PrintingWriter) Write(b []byte) (int, error) {
	s := string(b)
	fmt.Print(s)
	return len(b), nil
}

// called only in the forked child
func do_stuff_after_fork() error {
	new_root := "/srv/deb"
	var err error

	/*
		err := syscall.Mount(new_root, new_root, "", syscall.MS_BIND|syscall.MS_REC, "")
		if err != nil {
			return err
		}
	*/
	op := fmt.Sprintf("/.pivot_root_%d_%d", os.Getpid(), time.Now().Unix())
	new_root_op := new_root + op
	err = os.Mkdir(new_root_op, 0777)
	if err != nil {
		return err
	}
	err = syscall.PivotRoot(new_root, new_root_op)
	if err != nil {
		os.RemoveAll(new_root_op)
		return fmt.Errorf("pivot_root(\"%s\",\"%s\") failed: %s", new_root, new_root_op, err)
	}
	os.RemoveAll(op)
	return nil
}

func deser(s string, res *StartupState) error {
	by, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		Errorf(`failed base64 Decode`, err)
		return err
	}
	b := bytes.Buffer{}
	b.Write(by)
	d := gob.NewDecoder(&b)
	err = d.Decode(res)
	if err != nil {
		Errorf(`failed gob Decode`, err)
		return err
	}
	return nil
}
func ser(res *StartupState) (string, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(res)
	if err != nil {
		Errorf(`failed gob Encode`, err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}
