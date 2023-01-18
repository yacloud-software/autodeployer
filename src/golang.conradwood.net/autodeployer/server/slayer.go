package main

import (
	"fmt"
	"golang.conradwood.net/autodeployer/killer"
	"golang.conradwood.net/go-easyops/linux"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// kill all processes
func slayAll() {
	if isTestMode() {
		fmt.Printf("Not slaying - testmode activated\n")
		return
	}
	processChangeLock.Lock()
	defer processChangeLock.Unlock()
	users := getListOfUsers()
	var wg sync.WaitGroup
	for _, un := range users {
		wg.Add(1)
		go func(user string) {
			defer wg.Done()
			slay(user, *start_brutal)
		}(un.Username)
	}
	wg.Wait()
	// now find /srv/autodeployer/deployments binaries (in case they ran as root)
	pids, err := linux.AllPids()
	if err != nil {
		fmt.Printf("Failed to get all pids!! (%s)\n", err)
		return
	}
	for _, p := range pids {
		if !strings.HasPrefix(p.Binary(), "/srv/autodeployer/deployments") {
			continue
		}
		fmt.Printf("Killing %s\n", p)
		killer.KillPID(p.Pid(), false)

	}

}
func slay(username string, quick bool) {
	if username == "root" {
		panic("will not slay root")
	}
	var cmd []string
	su := sucom()
	kill := killcom()
	// we clean up - to make sure we really really release resources, we "slay" the user
	fmt.Printf("Slaying process of user %s (quick=%v)...\n", username, quick)
	if quick {
		cmd = []string{su, "-m", username, "-c", kill + ` -KILL -1`}
	} else {
		xcmd := exec.Command(su, "-m", username, "-c", kill+` -TERM -1`)
		err := xcmd.Run()
		if err != nil {
			fmt.Printf("failed to kill %s: %s\n", username, err)
		}
		time.Sleep(time.Duration(2) * time.Second)
		cmd = []string{su, "-m", username, "-c", kill + ` -KILL -1`}

	}
	if *debug {
		fmt.Printf("Command: %v\n", cmd)
	}
	l := linux.New()
	l.SetAllowConcurrency(true)
	l.SetMaxRuntime(time.Duration(15) * time.Second) // slay might sleep for 10 seconds between signals
	out, err := l.SafelyExecute(cmd, nil)
	if (*debug) && (err == nil) {
		fmt.Printf("Slayed process of user %s\n", username)
	}
	if err != nil {
		fmt.Printf("Command output: \n%s\n", out)
		fmt.Printf("su/kill -1 user %s failed: %s (trying slay executable)\n", username, err)
		if quick {
			cmd = []string{"slay", "-9", username}
		} else {
			cmd = []string{"slay", "-clean", username}
		}
		l := linux.New()
		l.SetAllowConcurrency(true)
		l.SetMaxRuntime(time.Duration(15) * time.Second) // slay might sleep for 10 seconds between signals
		out, err := l.SafelyExecute(cmd, nil)
		if err != nil {
			fmt.Printf("slay failed: %s\n%s", err, out)
		} else {
			fmt.Printf("fallback method using slay suceeded for user %s\n", username)
		}
	}
	setDeploymentsGauge()
}
