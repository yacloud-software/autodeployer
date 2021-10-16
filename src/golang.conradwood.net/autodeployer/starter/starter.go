package starter

import (
	"flag"
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/tokens"
	"golang.conradwood.net/go-easyops/utils"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	download_timeout = flag.Int("download_server_timeout", 300, "maximum number of `seconds` to download from server before timing out")
	message_id       string
	adc              ad.AutoDeployerClient
	memlimit         uint32 // the ACTUAL memlimit of our process (after adjusting for defaults/missing values)
	startupResponse  *ad.StartupResponse
)

// this is the non-privileged section of the autodeployer

//*********************************************************************
// execute whatever passed in as msgid and never returns
// (exits if childprocess exits)
// this is the 2nd part of the server (execed by main, this part is running unprivileged)
func Execute(mid string, autodeployer_port int) {
	message_id = mid
	// redirect stderr to stdout (to capture panics)
	syscall.Dup2(int(os.Stdout.Fd()), int(os.Stderr.Fd()))

	// we're speaking to the local server only ever
	serverAddr := fmt.Sprintf("localhost:%d", autodeployer_port)
	creds := client.GetClientCreds()
	fmt.Printf("Connecting to local autodeployer server:%s...\n", serverAddr)
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		fmt.Printf("fail to dial: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("Creating client...")
	adc = ad.NewAutoDeployerClient(conn)
	ctx := tokens.ContextWithToken()

	// the the server we're starting to deploy and get the parameters for deployment
	pid := os.Getpid()
	sr := ad.StartupRequest{Msgid: message_id, PID: uint64(pid)}
	srp, err := adc.InternalStartup(ctx, &sr)
	if err != nil {
		fmt.Printf("Failed to startup: %s\n", utils.ErrorString(err))
		os.Exit(10)
	}
	if srp.URL == "" {
		fmt.Printf("no download url in startup response\n")
		os.Exit(10)
	}
	if srp.AppReference == nil {
		fmt.Printf("Error: Application Definition is nil (%s)", srp.URL)
		os.Exit(10)
	}
	if srp.AppReference.AppDef == nil {
		fmt.Printf("Error: Application Definition is nil (%s)", srp.URL)
		os.Exit(10)
	}
	startupResponse = srp
	// set memory limits to something sane
	memlimit = 3000 // a default if nobody tells us otherwise
	rl := srp.Limits
	if rl != nil && rl.MaxMemory != 0 {
		memlimit = rl.MaxMemory
	}

	/*
		// only needed for oracle - openjdk just does The Right Thing
				if memlimit < 3000 && srp.AppReference.AppDef.Java {
					fmt.Printf("Memory limit of %d Megabytes is too small for java. Java needs at least 3000 Megabytes\n", memlimit)
					os.Exit(10)
				}
	*/
	if srp.UseSetRLimit {
		var rLimit syscall.Rlimit
		rLimit.Cur = uint64(memlimit) * 1024 * 1024
		rLimit.Max = uint64(memlimit) * 1024 * 1024
		err = syscall.Setrlimit(syscall.RLIMIT_AS, &rLimit)
		utils.Bail(fmt.Sprintf("Failed to set rlimit (%d megabytes)", memlimit), err)
	}
	if rl != nil {
		err = syscall.Setpriority(syscall.PRIO_PROCESS, pid, int(rl.Priority))
		utils.Bail(fmt.Sprintf("Failed to set prio of pid=%d to %d", pid, rl.Priority), err)
		fmt.Printf("Set process %d priority to %d\n", pid, rl.Priority)
	}

	unix.Umask(000)
	// change to my working directory
	err = os.Chdir(srp.WorkingDir)
	if err != nil {
		fmt.Printf("Failed to Chdir() to %s: %s\n", srp.WorkingDir, err)
	}
	fmt.Printf("Chdir() to %s\n", srp.WorkingDir)
	// download the binary and/or archive
	fmt.Printf("Downloading binary from %s\n", srp.URL)
	ctr := 0
	for {
		err = DownloadBinary(srp, ctr)
		if err == nil {
			break
		}
		ctr++
		if ctr > 3 {
			fmt.Printf("Failed to download from %s: %s\n", srp.URL, err)
			os.Exit(10)
		}
	}

	// execute the binary
	ports := countPortCommands(srp.Args)

	fmt.Printf("Getting resources\n")
	ctx = tokens.ContextWithToken()
	resources, err := adc.AllocResources(ctx, &ad.ResourceRequest{Msgid: message_id, Ports: int32(ports)})
	if err != nil {
		fmt.Printf("Failed to alloc resources: %s\n", err)
		os.Exit(10)
	}
	fmt.Printf("Start commandline: %s %v (%d ports)\n", srp.Binary, srp.Args, ports)
	rArgs := replacePorts(srp.Args, resources.Ports)
	rArgs = replaceSecureArgs(rArgs, srp.SecureArgs)

	cmd := createCommand(srp, rArgs)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Failed to start(): %s\n", err)
		os.Exit(10)
	}
	_, err = adc.Started(ctx, &ad.StartedRequest{Msgid: message_id, Args: rArgs})
	if err != nil {
		fmt.Printf("Failed to inform daemon about pending startup. aborting. (%s)\n", err)
		os.Exit(10)
	}
	err = cmd.Wait()
	if err == nil {
		fmt.Printf("Command completed with no error\n")
	} else {
		fmt.Printf("Command completed: %s\n", err)
	}
	failed := err != nil
	adc.Terminated(ctx, &ad.TerminationRequest{Msgid: message_id, Failed: failed})
	os.Exit(0)
}

//*********************************************************************
//*********************************************************************

// replace ${SECURE-XXXXX} with actual ports
func replaceSecureArgs(args []string, secargs map[string]string) []string {
	if secargs == nil || len(secargs) == 0 {
		fmt.Printf("No secure args\n")
		return args
	}
	fmt.Printf("got %d secure args\n", len(secargs))
	var res []string
	for _, a := range args {
		s := a
		//		fmt.Printf("Secure arg? \"%s\"\n", s)
		for k, v := range secargs {
			repl := fmt.Sprintf("${SECURE-%s}", k)
			ns := strings.ReplaceAll(s, repl, v)
			if s != ns {
				//				fmt.Printf("Replaced \"%s\"\n", repl)
				s = ns
				break
			}
		}
		res = append(res, s)

	}
	return res
}

// replace ${PORTx} with actual ports
func replacePorts(args []string, ports []int32) []string {
	res := []string{}
	for _, r := range args {
		n := r
		for i := 0; i < len(ports); i++ {
			s := fmt.Sprintf("${PORT%d}", i+1)
			n = strings.ReplaceAll(n, s, fmt.Sprintf("%d", ports[i]))
		}
		res = append(res, n)
	}
	return res
}

// how many ${PORTx} variables do we have in our string?
func countPortCommands(args []string) int {
	res := 0
	for i := 1; i < 20; i++ {
		s := fmt.Sprintf("${PORT%d}", i)
		for _, r := range args {
			if strings.Contains(r, s) {
				res = res + 1
			}
		}
	}
	return res
}

// download a file and depending on its type extract the archive
// this must either produce a "download.tar" or a file called "executable"
func DownloadBinary(srp *ad.StartupResponse, ctr int) error {
	outfile := srp.Binary
	archive := false
	if strings.HasSuffix(srp.URL, ".tar") {
		// unload an archive
		outfile = "download.tar"
		archive = true
	}
	if strings.HasSuffix(srp.URL, ".tar.bz2") {
		// unload an archive
		outfile = "download.tar.bz2"
		archive = true
	}
	// try get it from local autodeployer (caching) first, fallback to http
	err := DownloadFromServer(srp, message_id, ctr)
	if err != nil {
		// no longer do direct http
		/*
			fmt.Printf("Failed to download from server (%s) - retrying direct\n", err)
			err := DownloadFromURL(srp.URL, outfile, srp.DownloadUser, srp.DownloadPassword)
		*/
		if err != nil {
			return err
		}
	}
	if archive {
		err := untar(outfile)
		if err != nil {
			return err
		}
	}
	return nil
}

// untar tar file. if tar file contains other tarfiles (in dist/), untar those too
func untar(filename string) error {
	// extract it...
	// (should use library not external tool!)
	paras := "-xf"
	if strings.HasSuffix(filename, ".bz2") {
		paras = paras + "j" // autounzip
	}
	cmd := exec.Command("/bin/tar", paras, filename)
	op, err := cmd.CombinedOutput()
	os.Remove(filename) // remove in any case, even if error
	if err != nil {
		fmt.Printf("tar output: %s\n", op)
		return err
	}
	tars, err := findTars(".")
	if err != nil {
		return err
	}
	for _, t := range tars {
		cmd := exec.Command("/bin/tar", "-xf", t)
		op, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("(%s) tar output: %s\n", t, op)
			return err
		}
	}
	return nil
}

func findTars(dir string) ([]string, error) {
	var res []string
	files, err := ioutil.ReadDir(dir)
	for _, f := range files {
		if !f.IsDir() {
			if strings.HasSuffix(f.Name(), ".tar") {
				res = append(res, dir+"/"+f.Name())
			}
			continue
		}
		nf, err := findTars(dir + "/" + f.Name())
		if err != nil {
			return nil, err
		}
		res = append(res, nf...)

	}
	return res, err
}

func DownloadFromServer(srp *ad.StartupResponse, id string, ctr int) error {
	filename := filepath.Base(srp.URL)
	fmt.Printf("Downloading to %s\n", filename)
	if strings.HasSuffix(srp.URL, ".tar") {
		filename = "download.tar"
	}
	if strings.HasSuffix(srp.URL, ".tar.bz2") {
		filename = "download.tar.bz2"
	}
	ctx := tokens.ContextWithTokenAndTimeout(uint64(*download_timeout))
	sr := &ad.StartedRequest{Msgid: message_id}
	if ctr > 0 {
		sr.DownloadFailed = true
	}
	srv, err := adc.Download(ctx, sr)
	if err != nil {
		return err
	}

	output, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error while creating", filename, "-", err)
		return err
	}
	defer output.Close()
	pg := utils.ProgressReporter{Prefix: "Starter"}
	for {
		bd, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		_, err = output.Write(bd.Data)
		if err != nil {
			return err
		}
		pg.Add(uint64(len(bd.Data)))
		pg.Print()
	}
	return nil
}

// create a command from startup response with these args
func createCommand(srp *ad.StartupResponse, rArgs []string) *exec.Cmd {
	var err error
	var fullb string
	path, err := filepath.Abs(".")
	utils.Bail("failed to get absolute path for current working directory", err)

	if srp.AppReference.AppDef.Java {
		fmt.Printf("Starting java class...\n")
		fullb = "/usr/bin/java"
		cp, err := buildClassPath(path)
		utils.Bail("failed to get classpath", err)
		rArgs = append([]string{
			"-cp",
			cp,
			srp.Binary,
		}, rArgs...) // prepend classpath and mainclass
		rArgs = append(getJavaMemoryArgs(), rArgs...) // prepend memory args
	} else {
		// executing a "normal" binary
		fullb = fmt.Sprintf("%s/%s", path, srp.Binary)
		err = os.Chmod(fullb, 0500)
		if err != nil {
			fmt.Printf("Failed to chmod %s: %s\n", fullb, err)
			os.Exit(10)
		}
	}

	uname, err := user.Current()
	if err != nil {
		fmt.Printf("Failed to get current user: %s\n", err)
		os.Exit(10)
	}
	prio := int32(0)
	if srp.Limits != nil {
		prio = srp.Limits.Priority
	}
	scmd := fullb
	var nArgs []string
	if prio != 0 {
		nArgs = []string{"-n", fmt.Sprintf("%d", prio)}
		nArgs = append(nArgs, fullb)
		nArgs = append(nArgs, rArgs...)
		fmt.Printf("Starting user application (as user %s)..\n", uname.Username)
		fmt.Printf("Starting binary \"nice\" with %d args:\n", len(nArgs))
		scmd = "/usr/bin/nice"
	} else {
		nArgs = rArgs
	}
	/*
		// do not do this, it prints the secureargs
				for _, s := range nArgs {
					fmt.Printf("Arg: \"%s\"\n", s)
				}
	*/
	cmd := exec.Command(scmd, nArgs...)
	/*
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid:     true,
			Setctty:    true,
			Foreground: true,
			Noctty:     true,
			Setpgid:    true,
		}
	*/
	return cmd
}

// look in all known paths for jars. gradle puts them all over the place...
func buildClassPath(path string) (string, error) {
	jars, err := searchFor(path, ".jar")
	if err != nil {
		return "", err
	}
	return strings.Join(jars, ":"), nil
}

// given a directory and a suffix, recursively traverses dirs and returns matching filenames
func searchFor(dir string, suffix string) ([]string, error) {
	fmt.Printf("Finding jars in dir \"%s\"\n", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, f := range files {
		if !f.IsDir() {
			if strings.HasSuffix(f.Name(), suffix) {
				res = append(res, dir+"/"+f.Name())
			}
			continue
		}
		fs, err := searchFor(dir+"/"+f.Name(), suffix)
		if err != nil {
			return nil, err
		}
		res = append(res, fs...)

	}
	return res, nil
}

func getJavaMemoryArgs() []string {
	if true {
		return []string{
			/*
				tls support is actually quite borked in java. That's probably why google
				used netty instead of the built-in ssl.
				Either way, the StatusHandler exposes via ssl, so to avoid it breaking,
				we have to disable tls1.3
				here is one bug report:
				https://bugs.openjdk.java.net/browse/JDK-8207009
			*/
			"-Djdk.tls.client.protocols=TLSv1,TLSv1.1,TLSv1.2",
			"-XX:+UnlockExperimentalVMOptions"}
	}
	maxheap := int(memlimit-2100) * 1024 * 1024
	maxheap = 4283750400
	initialheap := int(memlimit/2) * 1024 * 1024
	threadsize := 10 * 1024 * 1024
	initialheap = maxheap / 2

	res := []string{
		/*
			tls support is actually quite borked in java. That's probably why google
			used netty instead of the built-in ssl.
			Either way, the StatusHandler exposes via ssl, so to avoid it breaking,
			we have to disable tls1.3
			here is one bug report:
			https://bugs.openjdk.java.net/browse/JDK-8207009
		*/
		"-Djdk.tls.client.protocols=TLSv1,TLSv1.1,TLSv1.2",
		"-XX:+PrintFlagsFinal",
		fmt.Sprintf("-Xmx%d", maxheap),
		fmt.Sprintf("-Xms%d", initialheap),
		fmt.Sprintf("-Xss%d", threadsize),
	}
	return res
}
