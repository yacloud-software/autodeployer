package main

import (
	"fmt"
	pb "golang.conradwood.net/apis/autodeployer"
	rpb "golang.conradwood.net/apis/registry"
	"golang.conradwood.net/autodeployer/cgroups"
	"golang.conradwood.net/autodeployer/config"
	"golang.conradwood.net/autodeployer/deployments"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/server"
	"golang.conradwood.net/go-easyops/tokens"
	"time"
)

// subset of deployments.Deployed

/*
 code that cleans up once the process exited
 this code runs in the autodeployer-server process
 it is - as supposed to the child process - highly likely
 that it is run each time the processes terminates since it
 is trigger by process exit (SIGCHLD) rather than any notification.
 so even SIGKILL or SIGSEGV will end up here
*/
var (
	regClient rpb.RegistryClient
)

// called by the main thread (privileged) when our forked startup.go finished
// (when the application exited)
func StartupCodeFinished(du *deployments.Deployed, exitCode error) {
	if regClient == nil {
		regClient = client.GetRegistryClient()
	}
	fmt.Printf("child %s terminated\n", du.StartupMsg)
	cgroups.RemoveCgroup(du)
	ctx := tokens.ContextWithToken()
	ds := &rpb.DeregisterServiceRequest{ProcessID: du.StartupMsg}
	_, err := regClient.V2DeregisterService(ctx, ds)
	if err != nil {
		fmt.Printf("Failed to deregister service %s: %s\n", du.String(), err)
	} // unregister the ports...
	err = server.UnregisterPortRegistry(du.Ports)
	if err != nil {
		fmt.Printf("Failed to unregister port %s\n", err)
	}
	du.Finished = time.Now()
	du.Status = pb.DeploymentStatus_TERMINATED
	if du.ExitCode == nil {
		du.ExitCode = exitCode
	}
	s := ""
	if du.ExitCode != nil {
		du.Log("Failed: %s (%s)", du.String(), du.ExitCode)
	} else {
		du.Log("Exited normally: %s", du.String())
	}

	config.AppStopped()
	Slay(du.User.Username, *brutal)

	if du.Logger != nil {
		du.Logger.LogCommandStdout(s, fmt.Sprintf("%s", du.Status))
		du.Logger.Close(du.GetExitCode())
	}
	setDeploymentsGauge()
}
