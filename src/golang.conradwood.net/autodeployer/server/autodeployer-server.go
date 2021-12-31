package main

// this is the main, privileged daemon. got to run as root because we're forking off
// different users from here

// fileaccess is split out to starter.go, which runs as an unprivileged user

// (this is done by virtue of exec('su',Args[0]) )
// the flag msgid goes into the startup code, so do not run the privileged daemon with that flag!

import (
	"context"
	"errors"
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/autodeployer"
	"golang.conradwood.net/apis/common"
	dm "golang.conradwood.net/apis/deploymonkey"
	rpb "golang.conradwood.net/apis/registry"
	"golang.conradwood.net/apis/secureargs"
	"golang.conradwood.net/autodeployer/cgroups"
	"golang.conradwood.net/autodeployer/config"
	"golang.conradwood.net/autodeployer/deployments"
	"golang.conradwood.net/autodeployer/downloader"
	"golang.conradwood.net/autodeployer/starter"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/linux" // add busy gauge
	"golang.conradwood.net/go-easyops/logger"
	"golang.conradwood.net/go-easyops/prometheus"
	"golang.conradwood.net/go-easyops/server"
	"golang.conradwood.net/go-easyops/tokens"
	"golang.conradwood.net/go-easyops/utils"
	"google.golang.org/grpc"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// static variables for flag parser
var (
	max_users           = flag.Int("max_users", 0, "if non zero, limit number of users to manage to this number")
	print_to_stdout     = flag.Bool("print_to_stdout", false, "print commands stdout to autodeployer stdout as well as to the logservice")
	slay_corrupt_stdout = flag.Bool("slay_corrupt_stdout", true, "if true, slay commands with a lines on stdout that cannot be processed, e.g. too long")
	stophandler         = flag.Bool("activate_stop_handler", true, "activates the stop handler. however this prevents us from writing panic() to stdout")
	force_registry      = flag.String("force_registry", "", "if not empty, all -registry=xx parameters will be rewritten to this `hostname`")
	use_cgroups         = flag.Bool("use_cgroups", true, "use cgroups (instead of other mechanisms)")
	msgid               = flag.String("msgid", "", "A msgid indicating that we've been forked() and execing the command. used internally only")
	add_instance_id     = flag.Bool("add_instance_id", true, "pass ge_instance_id to each binary being started")
	max_downloads       = flag.Int("max_downloads", 3, "max simultanous deployments in status DOWNLOADING")
	deploymentsGauge    = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "autodeployer_deployments",
			Help: "V=1 UNIT=ops DESC=current count of deployments on this instance",
		})

	processChangeLock sync.Mutex
	port              = flag.Int("port", 4000, "The server port")
	test              = flag.Bool("test", false, "set to true if you testing the server")
	debug             = flag.Bool("debug", false, "lots of debug output")
	portLock          = new(sync.Mutex)
	idleReaper        = flag.Int("reclaim", 5, "Reclaim terminated user accounts after `seconds`")
	startTimeout      = flag.Int("start_timeout", 5, "timeout a deployment after `seconds`")
	machineGroup      = flag.String("machinegroup", "worker", "the group a specific machine is in")
	testfile          = flag.String("cfgfile", "", "config file (for testing)")
	waitdir           = flag.String("waitdir", "", "Block startup until this `directory` exists")
	brutal            = flag.Bool("brutal", false, "brutally kill processes (-9 immediately)")
	start_brutal      = flag.Bool("brutal_start", false, "brutally kill processes (-9 immediately) on STARTUP only")
	deplMonkey        dm.DeployMonkeyClient
)

func isTestMode() bool {
	if *testfile != "" {
		return true
	}
	return false
}

// callback from the compound initialisation
func st(server *grpc.Server) error {
	s := new(AutoDeployer)
	// Register the handler object
	pb.RegisterAutoDeployerServer(server, s)
	return nil
}

func stopping() {
	fmt.Printf("Shutdown request received, slaying everyone...\n")
	slayAll()
	fmt.Printf("Shutting down, slayed everyone...\n")
	os.Exit(0) // warning: this prevents us from printing panics to stdout
}
func main() {
	var err error
	flag.Parse() // parse stuff. see "var" section above
	fmt.Printf("Starting autodeployer...\n")
	// if file does not exist, this will do NOTHING,
	// thus save to leave it in here w/o switch
	err = config.Start()
	utils.Bail("Unable to process config", err)
	prometheus.MustRegister(deploymentsGauge)
	if *stophandler {
		fmt.Printf("Setting sigterm handler...\n")
		// catch ctrl-c (for system shutdown)
		// and signal child processes
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			fmt.Printf("SIGTERM received\n")
			stopping()
			os.Exit(0)
		}()
		defer stopping()
	}
	if *msgid != "" {
		/********************************** branch to client startup code **************/
		starter.Execute(*msgid, *port)
		os.Exit(10) // should never happen
	}

	downloader.Start(func() bool {
		return *cache_enabled
	})
	CacheStart()
	if *waitdir != "" {
		for !exists(*waitdir) {
			fmt.Printf("Waiting for %s\n", *waitdir)
			time.Sleep(time.Second * 1)
		}
	}
	fmt.Printf("Slaying all autodeployers currently active...\n")
	// we are brutal - if we startup we slay all deployment users
	slayAll()
	fmt.Printf("Slayed all...\n")

	fmt.Printf("Autodeployer ready.\n")

	go started()

	sd := server.NewServerDef()
	sd.Port = *port
	sd.Register = st
	err = server.ServerStartup(sd)
	if err != nil {
		fmt.Printf("failed to start server: %s\n", err)
	}
	fmt.Printf("Done\n")
	return

}
func GetDeployMonkeyClient() dm.DeployMonkeyClient {
	if deplMonkey != nil {
		return deplMonkey
	}
	res := dm.NewDeployMonkeyClient(client.Connect("deploymonkey.DeployMonkey"))
	deplMonkey = res
	return res
}

//*********************************************************************
// called just before registering at the registry
func started() {
	for {
		time.Sleep(time.Duration(3) * time.Second)
		ctx := tokens.ContextWithToken()
		_, err := GetDeployMonkeyClient().AutodeployerStartup(ctx, &common.Void{})
		if err == nil {
			break
		}
		fmt.Printf("Failed to inform deploymonkey of our startup: %s\n", err)
	}
	fmt.Printf("Deploymonkey was told about us starting up.\n")
}

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
			Slay(user, *start_brutal)
		}(un.Username)
	}
	wg.Wait()

}
func testing() {
	time.Sleep(time.Second * 1) // server needs starting up...
	ad := new(AutoDeployer)

	dp := pb.DeployRequest{
		DownloadURL:  "http://localhost/application",
		RepositoryID: 0,
		Binary:       "testapp",
		Args:         []string{"-port=${PORT1}", "-http_port=${PORT2}"},
		BuildID:      123,
	}
	dr, err := ad.Deploy(nil, &dp)
	if err != nil {
		fmt.Printf("Failed to deploy %s\n", err)
		os.Exit(10)
	}
	fmt.Printf("Deployed %v.\n", dr)
	fmt.Printf("Waiting forever...(testing a daemon)\n")
	select {}
}

/**********************************
* implementing the functions here:
***********************************/
type AutoDeployer struct {
}

func (s *AutoDeployer) StopAutodeployer(ctx context.Context, cr *common.Void) (*common.Void, error) {
	go func() {
		slayAll()
		ctx := tokens.ContextWithToken()
		GetDeployMonkeyClient().AutodeployerShutdown(ctx, &common.Void{})
		os.Exit(0)
	}()
	return &common.Void{}, nil
}
func (s *AutoDeployer) Deploy(ctx context.Context, cr *pb.DeployRequest) (*pb.DeployResponse, error) {
	defer setDeploymentsGauge()
	processChangeLock.Lock()
	defer processChangeLock.Unlock()
	if cr.DownloadURL == "" {
		return nil, errors.New("DownloadURL is required")
	}
	if cr.RepositoryID == 0 {
		return nil, errors.New("RepositoryID is required")
	}

	dls := deployments.DownloadingDeployments()
	if len(dls) >= *max_downloads {
		return nil, fmt.Errorf("too many deployments in progress already (%d >= %d)", len(dls), *max_downloads)
	}

	// if it's a webpackage, go easy
	if cr.DeployType == "webpackage" {
		err := DeployWebPackage(cr)
		if err != nil {
			return nil, err
		}
		resp := pb.DeployResponse{
			Success: true,
			Message: "OK",
			Running: false}
		return &resp, nil
	}
	if cr.DeployType == "staticfiles" {
		err := DeployStaticFiles(cr)
		if err != nil {
			return nil, err
		}
		resp := pb.DeployResponse{
			Success: true,
			Message: "OK",
			Running: false}
		return &resp, nil
	}

	// it's not a webpackage...

	for _, ar := range cr.AutoRegistration {
		_, err := convStringToApitypes(ar.ApiTypes)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to convert apitypes for %v: %s", ar, err))
		}
	}
	fmt.Printf("Deploying %d, Build %d\n", cr.RepositoryID, cr.BuildID)
	users := getListOfUsers()
	du := allocUser(users)
	if du == nil {
		fmt.Printf("allocUser returned no deployment entry ;(\n")
		return nil, errors.New("Failed to allocate a user. Server out of processes?")
	}
	du.Started = time.Now()
	du.DeployRequest = cr
	du.Args = cr.Args
	path := fmt.Sprintf("%s/%s/%d/%d", du.Namespace(), du.Groupname(), du.RepositoryID(), du.Build())
	du.Deploymentpath = path
	checkLogger(du)
	if du.AppReference() == nil {
		fmt.Printf("No appreference submitted by deploymonkey\n")
	} else {
		fmt.Printf("appreference submitted by deploymonkey: %#v\n", du.AppReference())
	}

	_, wd := filepath.Split(du.User.HomeDir)
	wd = fmt.Sprintf("/srv/autodeployer/deployments/%s", wd)

	du.Log("Deploying \"%d\" as user \"%s\" in %s", cr.RepositoryID, du.User.Username, wd)
	uid, _ := strconv.Atoi(du.User.Uid)
	gid, _ := strconv.Atoi(du.User.Gid)
	err := createWorkingDirectory(wd, uid, gid)
	if err != nil && (!isTestMode()) {
		du.Status = pb.DeploymentStatus_TERMINATED
		du.ExitCode = err
		du.Log("Failed to create working directory %s: %s", wd, err)
		return nil, err
	}
	du.Log("Creating startupid...")
	du.StartupMsg = RandomString(32)
	binname := os.Args[0]
	du.Log("Binary name (self): \"%s\"", binname)
	if binname == "" {
		return nil, errors.New("Failed to re-exec self. check startup path of daemon")
	}
	cmd := exec.Command(sucom(), "-s", binname, du.User.Username, "--",
		fmt.Sprintf("-token=%s", tokens.GetServiceTokenParameter()),
		fmt.Sprintf("-msgid=%s", du.StartupMsg))
	du.Log("Executing: %v", cmd)
	// fill our deploystatus with stuff
	// copy deployment request to deployment descriptor

	cmd.SysProcAttr = &syscall.SysProcAttr{
		//		Setsid:     true,
		//		Setctty:    true,
		//		Foreground: true,
		//		Noctty:     true,
		//		Setpgid:    true,
		//		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Cloneflags: syscall.CLONE_NEWNS,
	}
	du.Cmd = cmd
	du.WorkingDir = wd

	du.Status = pb.DeploymentStatus_STARTING
	du.Stdout, err = du.Cmd.StdoutPipe()
	if err != nil {
		s := fmt.Sprintf("Could not get cmd output: %s", err)
		du.Log(s)
		du.Idle = true
		return nil, errors.New(s)
	}
	du.Log("Starting Command: %s", du.String())
	du.Log("RepositoryID: %d", du.RepositoryID())
	err = cmd.Start()
	if err != nil {
		du.Log("Command: %v failed", cmd)
		fmt.Printf("fork() starter.go: %v failed", cmd)
		du.Idle = true
		return nil, err
	}
	// reap children...
	go waitForCommand(du)

	// now we need to wait for our internal startup message..
	sloop := time.Now()
	lastStatus := du.Status
	for {
		if du.Status != lastStatus {
			du.Log("Command %s changed status from %s to %s", du.String(), lastStatus, du.Status)
			lastStatus = du.Status
		}
		// wait
		if du.Status == pb.DeploymentStatus_TERMINATED {
			if du.ExitCode != nil {
				if du.LastLine == "" {
					return nil, du.ExitCode
				}
				txt := fmt.Sprintf("%s (%s)", du.ExitCode, du.LastLine)
				du.Log(txt)
				return nil, errors.New(txt)
			}
			resp := pb.DeployResponse{
				ID:      du.StartupMsg,
				Success: true,
				Message: "OK",
				Running: false}
			return &resp, nil
		} else if du.Status == pb.DeploymentStatus_EXECUSER || du.Status == pb.DeploymentStatus_CACHEDSTART || du.Status == pb.DeploymentStatus_DOWNLOADING {
			resp := pb.DeployResponse{
				ID:      du.StartupMsg,
				Success: true,
				Message: "OK",
				Running: true}
			return &resp, nil
		}
		if time.Since(sloop) > (time.Duration(*startTimeout) * time.Second) {
			return nil, errors.New(fmt.Sprintf("Timeout after %d seconds", *startTimeout))
		}
	}

}
func (s *AutoDeployer) Undeploy(ctx context.Context, cr *pb.UndeployRequest) (*pb.UndeployResponse, error) {
	fmt.Printf("Locking for exclusivity to undeploy %s...\n", cr.ID)
	processChangeLock.Lock()
	defer processChangeLock.Unlock()
	fmt.Printf("Locked for exclusivity...\n")
	if cr.ID == "" {
		return nil, errors.New("Undeployrequest requires id")
	}
	dep := entryByMsg(cr.ID)
	if dep == nil {
		dep = entryByDeplID(cr.ID)
		if dep == nil {
			return nil, errors.New(fmt.Sprintf("No deployment with id %s", cr.ID))
		}
	}
	dep.Log("Undeploy request received")
	fmt.Printf("Undeploy request received: %v", dep)
	sb := ""
	if cr.Block {
		sb = fmt.Sprintf("Shutting down (sync): %s\n", dep.String())
		fmt.Println(s)
		Slay(dep.User.Username, *brutal)
	} else {
		sb = fmt.Sprintf("Shutting down (async): %s\n", dep.String())
		fmt.Println(s)
		go Slay(dep.User.Username, *brutal)
	}
	dep.Log(sb)
	res := pb.UndeployResponse{}
	return &res, nil
}

/*****************************************************
** this is called by the starter.go
** after it has forked, dropped privileges, and
** before it exec's the application
*****************************************************/
func (s *AutoDeployer) InternalStartup(ctx context.Context, cr *pb.StartupRequest) (*pb.StartupResponse, error) {
	d := entryByMsg(cr.Msgid)
	if d == nil {
		return nil, errors.New("No such deployment")
	}
	if d.Status != pb.DeploymentStatus_STARTING {
		return nil, errors.New(fmt.Sprintf("Deployment in status %s not STARTING!", d.Status))
	}
	d.Pid = cr.PID
	fmt.Printf("autodeployer-server: received internal startup() for %s\n", d.String())
	sr := &pb.StartupResponse{
		URL:              d.DeployRequest.DownloadURL,
		Args:             d.Args,
		Binary:           d.Binary(),
		DownloadUser:     d.DeployRequest.DownloadUser,
		DownloadPassword: d.DeployRequest.DownloadPassword,
		WorkingDir:       d.WorkingDir,
		Limits:           d.Limits(),
		AppReference:     d.AppReference(),
		UseSetRLimit:     !*use_cgroups,
	}
	if *use_cgroups {
		err := cgroups.ConfigCGroup(d)
		if err != nil {
			fmt.Printf("Failed to configure cgroups: %s\n", err)
			return nil, err
		}
		fmt.Printf("Added task %s to cgroup %d\n", d.StartupMsg, d.Cgroup)
	}
	if *add_instance_id {
		sr.Args = append(sr.Args, fmt.Sprintf("-ge_instance_id=%s", cr.Msgid)) //startupmsg
	}
	// we only retrieve secure args if the parameters require it
	retr := false
	for _, s := range d.Args {
		if strings.Contains(s, "${SECURE-") {
			retr = true
			break
		}
	}
	if retr {
		fmt.Printf("Getting secure args...\n")
		nc := tokens.ContextWithToken() // use autodeployer context to get secure args
		v, err := secureargs.GetSecureArgsServiceClient().GetArgs(nc, &secureargs.GetArgsRequest{RepositoryID: d.RepositoryID()})
		if err != nil {
			fmt.Printf("Failed to get secure args: %s\n", utils.ErrorString(err))
			return nil, err
		}
		if v != nil {
			sr.SecureArgs = v.Args
			fmt.Printf("Got %d secureargs\n", len(sr.SecureArgs))
		}

	}
	if *force_registry != "" {
		for i, s := range sr.Args {
			if strings.HasPrefix(s, "-registry=") {
				sr.Args[i] = "-registry=" + *force_registry
			}
		}
	}
	// add some standard args (which we pass to ALL deployments)
	sr.Args = append(sr.Args, fmt.Sprintf("-ge_deployment_descriptor=V1:%s", d.Deploymentpath))
	sr.Args = append(sr.Args, "-AD_started_by_auto_deployer=true")
	return sr, nil
}

// triggered by the unprivileged startup code
func (s *AutoDeployer) Started(ctx context.Context, cr *pb.StartedRequest) (*pb.EmptyResponse, error) {
	du := entryByMsg(cr.Msgid)
	if du == nil {
		return nil, errors.New("No such deployment")
	}
	du.Status = pb.DeploymentStatus_EXECUSER
	du.ResolvedArgs = cr.Args
	StartupCodeExec(du)
	return &pb.EmptyResponse{}, nil
}

// triggered by the unprivileged startup code
func (s *AutoDeployer) Terminated(ctx context.Context, cr *pb.TerminationRequest) (*pb.EmptyResponse, error) {
	d := entryByMsg(cr.Msgid)
	if d == nil {
		return nil, errors.New("No such deployment")
	}
	if cr.Failed {
		d.Log("Child reports: %s failed.", d.String())
		d.ExitCode = errors.New("Unspecified OS Failure")
	} else {
		d.Log("Child reports: %s exited", d.String())
	}
	d.Finished = time.Now()
	d.Status = pb.DeploymentStatus_TERMINATED
	return &pb.EmptyResponse{}, nil
}

// async, whenever a process exits...
func waitForCommand(du *deployments.Deployed) {
	lineOut := new(LineReader)
	buf := make([]byte, 1024)
	for {
		ct, err := du.Stdout.Read(buf)
		if err != nil {
			if err != io.EOF {
				du.Log("Failed to read command output: %s", err)
			}
			break
		}
		if ct == 0 {
			du.Log("stdout returned 0 bytes read")
			break
		}
		err = lineOut.gotBytes(buf, ct)
		if err != nil {
			du.Log("Error reading stdout: %s", err)
			fmt.Printf("%s invalid stdout!\n", du.String())
			if *slay_corrupt_stdout {
				du.Log("Killed process because of corrupt stdout: %s", err)
				du.Log("Last corrupt line: \"%s\"\n", lineOut.getBuf())
				Slay(du.User.Username, true)
			}
			lineOut.clearBuf()
		}
		for {
			line := lineOut.getLine()
			if line == "" {
				break
			}
			checkLogger(du)
			if *print_to_stdout {
				fmt.Printf(">>>>COMMAND: %s: %s\n", du.String(), line)
			}
			du.Log(line)
			du.LastLine = line
		}
	}
	err := du.Cmd.Wait()
	fmt.Printf("terminated: %s\n", du.String())

	// here we end up when our command terminates. it's still the privileged
	// server
	StartupCodeFinished(du, err)

}
func Slay(username string, quick bool) {
	var cmd []string
	su := sucom()
	kill := killcom()
	// we clean up - to make sure we really really release resources, we "slay" the user
	if *debug {
		fmt.Printf("Slaying process of user %s (quick=%v)...\n", username, quick)
	}
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
	out, err := l.SafelyExecute(cmd, nil)
	if (*debug) && (err == nil) {
		fmt.Printf("Slayed process of user %s\n", username)
	}
	if err != nil {
		fmt.Printf("Command output: \n%s\n", out)
		fmt.Printf("Slay user %s failed: %s\n", username, err)
	}
	setDeploymentsGauge()
}

// this is called by the starter
func (s *AutoDeployer) AllocResources(ctx context.Context, cr *pb.ResourceRequest) (*pb.ResourceResponse, error) {
	res := &pb.ResourceResponse{}
	d := entryByMsg(cr.Msgid)
	if d == nil {
		return nil, errors.New("No such deployment")
	}
	d.Status = pb.DeploymentStatus_RESOURCING
	fmt.Printf("Going into singleton port lock...\n")
	portLock.Lock()
	for i := 0; i < int(cr.Ports); i++ {
		res.Ports = append(res.Ports, allocPort(d))
	}
	portLock.Unlock()
	fmt.Printf("Done singleton port lock...\n")

	// tell the registry we're starting something!
	if regClient == nil {
		regClient = client.GetRegistryClient()
	}
	ctx = tokens.ContextWithToken() // we use the autodeployer token to create a service
	csr := &rpb.CreateServiceRequest{
		ProcessID:  d.StartupMsg,
		Partition:  "",
		DeployInfo: d.DeployInfo(),
	}
	fmt.Printf("Creating service %s with deployinfo: %s\n", d.String(), d.DeployInfoString())
	_, err := regClient.V2CreateService(ctx, csr)
	if err != nil {
		fmt.Printf("Failed to create service at registry: %s\n", utils.ErrorString(err))
		return nil, err
	}

	return res, nil
}

// we assume stuff is locked !
func allocPort(du *deployments.Deployed) int32 {
	startPort := 4100
	endPort := 4499
	for i := startPort; i < endPort; i++ {
		if !isPortInUse(i) {
			du.Ports = append(du.Ports, i)
			return int32(i)
		}
	}
	return 0
}
func isPortInUse(port int) bool {
	for _, d := range deployments.Deployments() {
		if d.Idle {
			continue
		}
		for _, p := range d.Ports {
			if p == port {
				return true
			}
		}

	}
	return false
}
func (s *AutoDeployer) ClearActions(ctx context.Context, req *common.Void) (*common.Void, error) {
	config.ClearApplied()
	return &common.Void{}, nil
}

func (s *AutoDeployer) GetDeployments(ctx context.Context, cr *pb.InfoRequest) (*pb.InfoResponse, error) {
	res := pb.InfoResponse{}
	for _, d := range deployments.Deployments() {
		if (d.Status == pb.DeploymentStatus_TERMINATED) || (d.Idle) {
			continue
		}
		da := d.DeployedApp()
		secArgs(ctx, da)
		res.Apps = append(res.Apps, da)
	}
	return &res, nil
}

func (s *AutoDeployer) GetMachineInfo(ctx context.Context, cr *pb.MachineInfoRequest) (*pb.MachineInfoResponse, error) {
	mg := strings.Split(*machineGroup, ",")
	var sx []string
	for _, m := range mg {
		if len(m) < 1 {
			continue
		}
		sx = append(sx, m)
	}
	if len(sx) == 0 {
		sx = []string{"worker"}
	}
	res := pb.MachineInfoResponse{MachineGroup: sx}
	return &res, nil
}

/**********************************
* implementing helper functions here
***********************************/
// given a user string will get the entry for that user
// will always return one (creates one if necessary)
func entryForUser(user *user.User) *deployments.Deployed {
	for _, d := range deployments.Deployments() {
		if d.User.Username == user.Username {
			return d
		}
	}

	// we create a new Deployed (for a given user)
	d := &deployments.Deployed{User: user, Idle: true, LastUsed: time.Now()}
	deployments.Add(d)
	return d
}

// find entry by msgid. nul if none found
func entryByMsg(msgid string) *deployments.Deployed {
	for _, d := range deployments.Deployments() {
		if d.StartupMsg == msgid {
			return d
		}
	}
	return nil
}

// find entry by deploymentid. nul if none found
func entryByDeplID(msgid string) *deployments.Deployed {
	for _, d := range deployments.Deployments() {
		if d.DeploymentID() == msgid {
			return d
		}
	}
	return nil
}

// given a list of users will pick one that is currently not used for deployment
// returns username
func allocUser(users []*user.User) *deployments.Deployed {
	for _, d := range deployments.Deployments() {
		freeEntry(d)
	}

	var lru *deployments.Deployed
	for _, u := range users {
		d := entryForUser(u)
		if d.Idle {
			if (lru == nil) || (lru.LastUsed.After(d.LastUsed)) {
				lru = d
			}
		}
	}
	if lru != nil {
		allocEntry(lru)
		return lru
	}
	fmt.Printf("Given %d users, found NO free entry\n", len(users))
	return nil
}

// frees an entry for later usage
func freeEntry(d *deployments.Deployed) {
	// it's already idle, nothing to do
	if d.Idle {
		return
	}
	// it's not idle and not terminated, so keep it!
	if d.Status != pb.DeploymentStatus_TERMINATED {
		return
	}

	// it's too soon after process terminated, we keep it around for a bit
	if time.Since(d.Finished) < (time.Duration(*idleReaper) * time.Second) {
		return
	}
	if d.Logger != nil {
		d.Logger.Close(d.GetExitCode())
		d.Logger = nil
	}

	os.RemoveAll(d.WorkingDir)
	fmt.Printf("Reclaimed %s\n", d.String())
	d.Idle = true
	// clear this to make really sure we're not reusing a logger
	d.DeployRequest = nil

}

// prepares an allocEntry for usage
func allocEntry(d *deployments.Deployed) {
	if d.Logger != nil {
		d.Logger.Close(d.GetExitCode())
		d.Logger = nil
	}
	d.LastUsed = time.Now()
	d.Idle = false
	d.Status = pb.DeploymentStatus_PREPARING
	// this is WAY too early (we haven't got appname and stuff yet)
	// this will reuse PREVIOUS log settings and thus confuse everyone terribly
	//	checkLogger(d)
}

func checkLogger(d *deployments.Deployed) {
	if d.Logger != nil {
		return
	}
	l, err := logger.NewAsyncLogQueue(d.Binary(), d.RepositoryID(), d.Groupname(), d.Namespace(), d.DeploymentID())
	if err != nil {
		fmt.Printf("Failed to initialize logger! %s\n", err)
	} else {
		d.Logger = l
	}
}

// creates a pristine, fresh, empty, standard nice working directory
func createWorkingDirectory(dir string, uid int, gid int) error {

	// we are going to delete the entire directory, so let's make
	// sure it's the right directory!
	if !strings.HasPrefix(dir, "/srv/autodeployer") {
		return errors.New(fmt.Sprintf("%s is not absolute", dir))
	}
	err := os.RemoveAll(dir)

	if err != nil {
		return errors.New(fmt.Sprintf("Failed to remove directory %s: %s", dir, err))
	}
	err = os.Mkdir(dir, 0700)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to mkdir %s: %s", dir, err))

	}
	f, err := os.Open(dir)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to open %s: %s", dir, err))
	}
	defer f.Close()
	err = f.Chown(uid, gid)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to chown %s: %s", dir, err))
	}
	return nil
}

// cycles through users deploy1, deploy2, deploy3 ... until the first one not found
func getListOfUsers() []*user.User {
	var res []*user.User
	i := 1
	for {
		if *max_users != 0 && len(res) >= *max_users {
			return res
		}
		un := fmt.Sprintf("deploy%d", i)
		u, err := user.Lookup(un)
		if err != nil {
			fmt.Printf("Max users: %d (ended with %s)\n", i, err)
			break
		}
		res = append(res, u)
		i++
	}
	return res
}

// called by the main thread, once the startup code claims handed control
// to the program
func StartupCodeExec(du *deployments.Deployed) {
	// auto register stuff
	fmt.Printf("Got %d autoregistration services to take care of\n", len(du.AutoRegistrations()))
	for _, ar := range du.AutoRegistrations() {
		port := du.GetPortByName(ar.Portdef)
		if port == 0 {
			fmt.Printf("Broken autoregistration (%v) - no portdef\n", ar)
			continue
		}
		apiTypes, err := convStringToApitypes(ar.ApiTypes)
		if err != nil {
			fmt.Printf("Broken autoregistration (%v) - api error %s\n", ar, err)
			continue
		}
		for _, at := range apiTypes {
			fmt.Printf("Autoregistering %s on port %d, type=%v\n", ar.ServiceName, port, at)
			if at == rpb.Apitype_tcp {
				sd := server.NewTCPServerDef(ar.ServiceName)
				sd.Port = port
				sd.DeployPath = du.Deploymentpath
				server.AddRegistry(sd)
			} else if at == rpb.Apitype_html {
				sd := server.NewHTMLServerDef(ar.ServiceName)
				sd.Port = port
				sd.DeployPath = du.Deploymentpath
				server.AddRegistry(sd)
			} else {
				fmt.Printf("Cannot (yet) auto-register apitype: %s\n", at)
				continue
			}

		}
	}
	if du.Logger != nil {
		du.Logger.SetStartupID(du.StartupMsg)
	}
	// application is running apparently
	config.AppStarted(du)
	setDeploymentsGauge()

}
func convStringToApitypes(apitypestring string) ([]rpb.Apitype, error) {
	var res []rpb.Apitype
	asa := strings.Split(apitypestring, ",")
	for _, as := range asa {
		as = strings.TrimLeft(as, " ")
		as = strings.TrimRight(as, " ")
		fmt.Printf("Converting: \"%s\"\n", as)
		v, ok := rpb.Apitype_value[as]
		if !ok {
			return nil, errors.New(fmt.Sprintf("unknown apitype \"%s\"", as))
		}
		av := rpb.Apitype(v)
		res = append(res, av)
	}
	return res, nil
}

// exists returns whether the given file or directory exists or not
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

// update the gauge for prometheus
func setDeploymentsGauge() {
	ct := deployments.ActiveDeployments()
	deploymentsGauge.Set(float64(len(ct)))
}

func secArgs(ctx context.Context, d *pb.DeployedApp) {
	if auth.IsRoot(ctx) {
		return
	}
	if d.Deployment == nil {
		d.Deployment.ResolvedArgs = nil
	}

}

func killcom() string {
	fs := []string{"/bin/kill", "/usr/bin/kill"}
	for _, f := range fs {
		if utils.FileExists(f) {
			return f
		}
	}
	panic("'su' not found")
}
func sucom() string {
	fs := []string{"/bin/su", "/usr/bin/su"}
	for _, f := range fs {
		if utils.FileExists(f) {
			return f
		}
	}
	panic("'su' not found")
}
