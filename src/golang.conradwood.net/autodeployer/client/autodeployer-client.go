package main

// instruct the autodeployer on a given server to download & deploy stuff

import (
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/autodeployer"
	cm "golang.conradwood.net/apis/common"
	dm "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/common"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/utils"
	"google.golang.org/grpc"
	"os"
	"sort"
	"strings"
)

// static variables for flag parser
var (
	pversion        = flag.Bool("print_version", false, "print autodeployer-server version")
	pmachineinfo    = flag.Bool("print_machineinfo", false, "print autodeployer-server machineinfo")
	stop            = flag.Bool("stop", false, "stop autodeployer")
	pkgname         = flag.String("package", "", "query or install a package")
	install_package = flag.Bool("install_package", false, "if true install package")
	details         = flag.Bool("details", false, "print details")
	server          = flag.String("server", "", "If provided, connects to a specific autodeployer (instead of lookup via registry)")
	clear           = flag.Bool("clear_actions", false, "clear list of applied actions on server")
	downloaduser    = flag.String("user", "", "the username to authenticate with at the downloadurl")
	downloadpw      = flag.String("password", "", "the password to authenticate with at the downloadurl")
	downloadurl     = flag.String("url", "", "The `URL` of the binary to deploy")
	binary          = flag.String("binary", "", "The relative path to the binary to deploy")
	paras           = flag.String("paras", "", "The parameters to pass to the binary")
	buildid         = flag.Int("build", 1, "The BuildID of the binary to be deployed")
	repo            = flag.Uint64("repo", 0, "The name of the repository where the source of the binary to be deployed lives.")
	group           = flag.String("group", "grp", "The name of the group")
	namespace       = flag.String("namespace", "namespc", "The namespace of the application")
	deployid        = flag.String("deploy_id", "", "an opaque token that is linked to this particular deployment (and returned in deploymentrequest")
	cfg             = flag.String("deploy", "", "a deploy.yaml `filename` to deploy")
	list            = flag.Bool("list", false, "list deployments")
	undeploy        = flag.Bool("undeploy", false, "undeploy [-deploy_id]")
	setlimits       = flag.Bool("set_limits", true, "set max runtime limits")
	maxmb           = flag.Int("max_mb", 1000, "runtime limit: maximum amount of memory (in megabytes) (see set_limits) ")
	cl              pb.AutoDeployerClient
	cache           = flag.String("cache", "", "download and cache a url")
)

func main() {
	flag.Parse()
	grpc.EnableTracing = true
	var conn *grpc.ClientConn
	var err error
	if *server != "" {
		s := *server
		if !strings.Contains(s, ":") {
			s = fmt.Sprintf("%s:4000", *server)
		}
		//fmt.Printf("Connecting to server %s\n", s)
		conn, err = grpc.Dial(s, grpc.WithTransportCredentials(client.GetClientCreds()))
		utils.Bail("Unable to connect to server", err)
	} else {
		conn = client.Connect("autodeployer.AutoDeployer")
	}

	cl = pb.NewAutoDeployerClient(conn)
	if *stop {
		utils.Bail(fmt.Sprintf("failed to stop: %s\n", *server), Stop())
		os.Exit(0)
	}
	if *pkgname != "" {
		utils.Bail("failed to process package", QueryOrInstallPackage())
		os.Exit(0)
	}
	if *cache != "" {
		Cache()
		os.Exit(0)
	}
	if *undeploy {
		Undeploy()
		os.Exit(0)
	}
	if *cfg != "" {
		deployFile()
		os.Exit(0)
	}
	if *clear {
		Clear()
		os.Exit(0)
	}
	if *pversion {
		utils.Bail("version failed", version())
		os.Exit(0)
	}
	if *pmachineinfo {
		utils.Bail("machineinfo failed", machineinfo())
		os.Exit(0)
	}
	if *list {
		listDeployments()
		os.Exit(0)
	}
	if *downloadurl != "" {
		deploy()
	}
}
func Clear() {
	ctx := authremote.Context()
	_, err := cl.ClearActions(ctx, &cm.Void{})
	utils.Bail("Failed to clear actions", err)
	fmt.Printf("Done")

}

func listDeployments() {
	ctx := authremote.Context()
	ir, err := cl.GetDeployments(ctx, &pb.InfoRequest{})
	utils.Bail("Failed to get deployments", err)
	fmt.Printf("%d deployments\n", len(ir.Apps))
	t := utils.Table{}
	t.AddHeaders("#", "AppID", "User", "Group", "DeploymentID", "BuildID", "Binary", "RepositoryID", "Status", "Since", "Pids", "CPids")
	sort.Slice(ir.Apps, func(i, j int) bool {
		return ir.Apps[i].Deployment.RepositoryID < ir.Apps[j].Deployment.RepositoryID
	})
	for i, app := range ir.Apps {
		di := app.Deployment
		rs := fmt.Sprintf("%d seconds", di.RuntimeSeconds)
		t.AddInt(i)
		t.AddString(app.ID)
		t.AddString(di.UserID)
		t.AddString(di.GroupID)
		t.AddString(di.DeploymentID)
		t.AddUint64(di.BuildID)
		t.AddString(di.Binary)
		t.AddUint64(di.RepositoryID)
		t.AddString(fmt.Sprintf("%v", di.Status))
		t.AddString(rs)
		t.AddUint64(di.Pid)
		s := ""
		for _, p := range di.ChildPids {
			s = s + fmt.Sprintf("%d ", p)
		}
		t.AddString(s)
		if *details {
			s := ""
			as := di.Args
			if len(di.ResolvedArgs) != 0 {
				as = di.ResolvedArgs
			}
			for _, a := range as {
				s = s + fmt.Sprintf("   %s", a)
			}
			t.AddString(s)
		}
		t.NewRow()
	}
	fmt.Println(t.ToPrettyString())
}

func deploy() {
	ctx := authremote.Context()
	d := *deployid
	if d == "" {
		d = "TEST_DEPLOY_ID"
	}
	rl := dm.Limits{
		MaxMemory: uint32(*maxmb),
	}
	req := pb.DeployRequest{DownloadURL: *downloadurl,
		Binary:           *binary,
		BuildID:          uint64(*buildid),
		DownloadUser:     *downloaduser,
		DownloadPassword: *downloadpw,
		RepositoryID:     *repo,
		Groupname:        *group,
		Namespace:        *namespace,
		DeploymentID:     d}
	if *setlimits {
		req.Limits = &rl
	}
	if *paras != "" {
		args := strings.Split(*paras, " ")
		req.Args = args
	}

	for i, para := range req.Args {
		fmt.Printf("Arg #%d %s\n", i, para)
	}
	resp, err := cl.Deploy(ctx, &req)
	if err != nil {
		fmt.Printf("Failed to deploy %d-%d from %s: %s\n", req.RepositoryID, req.BuildID, req.DownloadURL, err)
		return
	}
	fmt.Printf("Response to deploy: %v\n", resp)
}

func deployFile() {
	ctx := authremote.Context()
	d := *deployid
	if d == "" {
		d = "fake-deploymentid"
	}
	ar := &dm.AppReference{ID: 123456}
	fdef, err := common.ParseFile(*cfg, *repo)
	utils.Bail("failed to parse file", err)
	fmt.Printf("File: %v\n", fdef)
	for _, gd := range fdef.Groups {
		for _, app := range gd.Applications {
			url := app.DownloadURL
			if *downloadurl != "" {
				url = *downloadurl
			}
			ar.AppDef = app
			url = strings.Replace(url, "${BUILDID}", "latest", -1)
			req := pb.DeployRequest{
				DownloadURL:      url,
				Binary:           app.Binary,
				BuildID:          uint64(*buildid),
				DownloadUser:     *downloaduser,
				DownloadPassword: *downloadpw,
				RepositoryID:     app.RepositoryID,
				Groupname:        gd.GroupID,
				Namespace:        gd.Namespace,
				DeploymentID:     d,
				Args:             app.Args,
				AppReference:     ar,
			}

			resp, err := cl.Deploy(ctx, &req)
			utils.Bail("Failed to deploy", err)
			fmt.Printf("Response to deploy: %v\n", resp)
		}
	}
}

func Undeploy() {
	if *deployid == "" {
		fmt.Printf("Missing deployid\n")
		os.Exit(10)
	}
	ctx := authremote.Context()
	_, err := cl.Undeploy(ctx, &pb.UndeployRequest{ID: *deployid, Block: true})
	utils.Bail("Failed to undeploy", err)
	fmt.Printf("Done\n")

}

func Cache() {
	cr := &pb.URLRequest{URL: *cache}
	ctx := authremote.Context()
	r, err := cl.CacheURL(ctx, cr)
	utils.Bail("failed to cache url", err)
	p := float64(r.BytesDownloaded) / float64(r.TotalBytes) * 100
	fmt.Printf("Downloaded %d of %d bytes (%0.1f%%)\n", r.BytesDownloaded, r.TotalBytes, p)

}

func QueryOrInstallPackage() error {
	pname := *pkgname
	inst := *install_package

	var res *pb.PackageInstallResponse
	var err error
	ctx := authremote.Context()
	req := &pb.PackageInstallRequest{Name: pname}
	if inst {
		fmt.Printf("Installing package \"%s\"\n", pname)
		res, err = cl.InstallPackage(ctx, req)
	} else {
		fmt.Printf("Querying package \"%s\"\n", pname)
		res, err = cl.CheckPackage(ctx, req)
	}
	if err != nil {
		return err
	}
	fmt.Printf("Package \"%s\" installed: %v\n", res.Name, res.Installed)

	return nil
}
func Stop() error {
	ctx := authremote.Context()
	_, err := cl.StopAutodeployer(ctx, &pb.StopRequest{})
	return err
}
func version() error {
	ctx := authremote.Context()
	v, err := cl.GetMachineInfo(ctx, &pb.MachineInfoRequest{})
	if err != nil {
		return err
	}
	fmt.Printf("%d\n", v.AutoDeployerVersion)
	return nil
}

// suitable for reading in bash like so:
// read -r version id secs <<<$(autodeployer-client -pmachine_info) ; echo "$version|$id|$secs"
func machineinfo() error {
	ctx := authremote.Context()
	v, err := cl.GetMachineInfo(ctx, &pb.MachineInfoRequest{})
	if err != nil {
		return err
	}
	fmt.Printf("%d\n", v.AutoDeployerVersion)
	fmt.Printf("%s\n", v.InstanceID)
	fmt.Printf("%d\n", v.SecondsRunning)
	return nil
}
