package main

// instruct the autodeployer on a given server to download & deploy stuff

import (
	"errors"
	"flag"
	"fmt"
	"golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/deploymonkey"
	dc "golang.conradwood.net/deploymonkey/common"
	"golang.conradwood.net/deploymonkey/config"
	"golang.conradwood.net/deploymonkey/suggest"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/utils"
	"google.golang.org/grpc"
	"os"
	"strings"
	"time"
)

// static variables for flag parser
var (
	continue_on_error = flag.Bool("continue_on_error", false, "if true ignore deploy errors and continue to next deployment")
	del_version       = flag.Uint64("delete_version", 0, "if not 0, undeploy and delete this version")
	depllocal         = flag.Bool("deploy_local", false, "deploy the files in current git repository on local machine (using deploy.yaml in current directory)")
	depllist          = flag.Bool("deployments", false, "list current deployments")
	list_deployers    = flag.Bool("deployers", false, "list known autodeployers")
	dosuggest         = flag.Bool("suggest", false, "suggest fixes")
	applysuggest      = flag.Bool("apply_suggest", false, "suggest & applyfixes")
	short             = flag.Bool("short", false, "short listing")
	filename          = flag.String("configfile", "", "the yaml config file to submit to server")
	namespace         = flag.String("namespace", "", "namespace of the group to update")
	groupname         = flag.String("groupname", "", "groupname of the group to update")
	repository        = flag.Uint64("repository", 0, "repository of the app in the group to update")
	buildid           = flag.Int("buildid", 0, "the new buildid of the app in the group to update")
	binary            = flag.String("binary", "", "the binary of the app in the group to update")
	apply_version     = flag.Int("apply_version", 0, "(re-)apply a given version (expects `versionid`)")
	applyall          = flag.Bool("apply_all", false, "reapply ALL groups")
	applypending      = flag.Bool("apply_pending", false, "reapply any pending group versions")
	list              = flag.String("list", "", "list this `repository` previous versions")
	deployers         = flag.Bool("list_deployers", false, "if true list all known autodeployers")
	undeploy_app      = flag.Int("undeploy_version", 0, "undeploy applications of a given version (expects `versionid`)")
	print_sample      = flag.Bool("print_sample", false, "print a sample deploy.yaml")
	depl              pb.DeployMonkeyClient
)

func main() {
	flag.Parse()

	fmt.Printf("Starting deploymonkey client...\n")
	depl = pb.NewDeployMonkeyClient(client.Connect("deploymonkey.DeployMonkey"))

	done := false
	if *del_version != 0 {
		utils.Bail("failed to delete version", delVersion())
		os.Exit(0)
	}
	if *deployers {
		listDeployers()
		os.Exit(0)
	}
	if *print_sample {
		dc.PrintSample()
		os.Exit(0)
	}

	if *depllocal {
		DeployLocal()
		os.Exit(0)
	}

	if *depllist {
		listDeployments()
		os.Exit(0)
	}
	if *dosuggest {
		listSuggestions()
		os.Exit(0)
	}
	if *applysuggest {
		applySuggestions()
		os.Exit(0)
	}

	if *undeploy_app != 0 {
		undeployApplication(*undeploy_app)
		os.Exit(0)
	}
	if *list != "" {
		callListVersions(*list)
		os.Exit(0)
	}

	if *apply_version != 0 {
		applyVersion()
		done = true
	}
	if *filename != "" {
		processFile()
		done = true
		os.Exit(0)
	}
	if *namespace != "" {
		if *binary != "" {
			updateApp()
		} else {
			updateRepo()
		}
		done = true
	}
	if !done {
		listConfig()
		fmt.Printf("Nothing to do.\n")
		os.Exit(1)
	}
	os.Exit(0)
}
func bail(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Printf("%s: %s\n", msg, err)
	os.Exit(10)
}

func undeployApplication(ID int) {
	ctx := authremote.Context()
	uar := pb.UndeployApplicationRequest{ID: int64(ID)}
	resp, err := depl.UndeployApplication(ctx, &uar)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	if resp.App == nil {
		fmt.Printf("Nothing undeployed! Got the right ID?\n")
	} else {
		fmt.Printf("Undeployed: %d [%s]\n", resp.App.RepositoryID, resp.App.Binary)
	}
	for _, host := range resp.Host {
		fmt.Printf(" on %s\n", host)
	}
}

func callListVersions(repo string) {
	ctx := authremote.Context()
	dr := pb.ListVersionByNameRequest{Name: repo}
	resp, err := depl.ListVersionsByName(ctx, &dr)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	fmt.Printf("%d apps\n", len(resp.Apps))
	for i, a := range resp.Apps {
		if i > 10 {
			break
		}
		created := time.Unix(a.Created, 0)
		fmt.Printf("Version #%d: created %v, Build %d, binary %s\n", a.VersionID, created, a.Application.BuildID, a.Application.Binary)
	}
}

func applyVersion() {
	fmt.Printf("Applying Version...\n")
	ctx := authremote.Context()

	dr := pb.DeployRequest{VersionID: fmt.Sprintf("%d", *apply_version)}
	resp, err := depl.DeployVersion(ctx, &dr)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	fmt.Printf("Version %d applied (%s)\n", *apply_version, resp)

}

/************************************************
* this is more or less a direct copy of GetConfig
* and should be converted to use GetConfig()
*************************************************/
func listConfig() {
	config, err := config.GetConfig(depl)
	utils.Bail("failed to get config", err)
	for _, n := range config.Namespaces() {
		if !matchesArgs(n) {
			continue
		}
		gns := config.Groups(n)
		if (*short) && len(gns) == 1 {
			fmt.Printf("  %s (%d groups) - pending_versionID: #%d\n", n, len(gns), gns[0].PendingVersion)
			continue
		}
		fmt.Printf("  %s (%d groups) pending_version=%d\n", n, len(gns), gns[0].PendingVersion)
		for _, gs := range gns {
			gapps := config.Apps(gs.NameSpace, gs.GroupID)
			marker := ""
			if gs.PendingVersion != gs.DeployedVersion {
				marker = " ** <-- **"
			}
			fmt.Printf("      %s (%d applications)%s\n", gs, len(gapps), marker)
			for _, app := range gapps {
				ao := "__"
				if app.AlwaysOn {
					ao = "AO"
				}
				crit := "__"
				if app.Critical {
					crit = "CR"
				}
				fmt.Printf("           [%s%s] %dx App ID=%d, Repo=%d, Binary=%s, BuildID=#%d, %d autoregistrations\n", ao, crit, app.Instances, app.ID, app.RepositoryID, app.Binary, app.BuildID, len(app.AutoRegs))
			}
		}
	}
}
func matchesArgs(namespace string) bool {
	args := flag.Args()
	if len(args) == 0 {
		return true
	}
	for _, s := range args {
		if strings.Contains(namespace, s) {
			return true
		}
	}
	return false
}

func updateRepo() {
	if *namespace == "" {
		bail(errors.New("Namespace required"), "Cannot update repo")
	}
	if *groupname == "" {
		bail(errors.New("Groupname required"), "Cannot update repo")
	}
	if *repository == 0 {
		bail(errors.New("Repository required"), "Cannot update repo")
	}
	if *buildid == 0 {
		bail(errors.New("BuildID required"), "Cannot update repo")
	}
	fmt.Printf("Updating all apps in repo %d in group %s to buildid %d\n", *repository, *groupname, *buildid)
	ur := pb.UpdateRepoRequest{
		Namespace:    *namespace,
		GroupID:      *groupname,
		RepositoryID: *repository,
		BuildID:      uint64(*buildid),
	}
	ctx := authremote.Context()
	resp, err := depl.UpdateRepo(ctx, &ur)
	bail(err, "Failed to update repo")
	fmt.Printf("Response to updaterepo: %v\n", resp)
	return
}

func updateApp() {
	ad := pb.ApplicationDefinition{
		RepositoryID: *repository,
		Binary:       *binary,
		BuildID:      uint64(*buildid),
	}
	uar := pb.UpdateAppRequest{
		GroupID:   *groupname,
		Namespace: *namespace,
		App:       &ad,
	}
	fmt.Printf("Updating app %s\n", *binary)
	ctx := authremote.Context()

	resp, err := depl.UpdateApp(ctx, &uar)
	if err != nil {
		fmt.Printf("Failed to update app: %s\n", err)
		return
	}
	fmt.Printf("Response to updateapp: %v\n", resp.Result)
}

func processFile() {
	fmt.Printf("Processing file...\n")
	if *namespace != "" {
		fmt.Printf("-configfile and -namespace are mutually exclusive\n")
		os.Exit(10)
	}
	fd, err := dc.ParseFile(*filename, *repository)
	if err != nil {
		fmt.Printf("Failed to parse file %s: %s\n", *filename, err)
		os.Exit(10)
	}
	// print limits...
	for _, group := range fd.Groups {
		for _, app := range group.Applications {
			app.BuildID = uint64(*buildid)
			fmt.Printf("Binary: %s\n", app.Binary)
			fmt.Printf("   Limits: %#v\n", app.Limits)
		}
	}

	*namespace = fd.Namespace
	fmt.Printf("Set namespace to \"%s\"\n", *namespace)
	grpc.EnableTracing = true
	ctx := authremote.Context()

	for _, req := range fd.Groups {
		resp, err := depl.DefineGroup(ctx, req)
		if err != nil {
			fmt.Printf("Failed to define group: %s\n", err)
			return
		}
		if resp.Result != pb.GroupResponseStatus_CHANGEACCEPTED {
			fmt.Printf("Response to deploy: %s - skipping\n", resp.Result)
			continue
		}
		dr := pb.DeployRequest{VersionID: resp.VersionID}
		dresp, err := depl.DeployVersion(ctx, &dr)
		if err != nil {
			fmt.Printf("Failed to deploy version %s: %s\n", resp.VersionID, err)
			return
		}
		fmt.Printf("Deploy response: %v\n", dresp)
	}
}

func listSuggestions() {
	ctx := authremote.Context()
	depls, err := depl.GetDeploymentsFromCache(ctx, &common.Void{})
	utils.Bail("Failed to get deployments from cache", err)
	cfg, err := config.GetConfig(depl)
	utils.Bail("Could not get config", err)
	s, err := suggest.Analyse(cfg, depls)
	utils.Bail("Suggestion failed", err)
	fmt.Println(s.String())
	if len(s.Starts) != 0 || len(s.Stops) != 0 {
		os.Exit(1)
	}
}
func applySuggestions() {
	ctx := authremote.Context()
	depls, err := depl.GetDeploymentsFromCache(ctx, &common.Void{})
	utils.Bail("Failed to get deployments from cache", err)
	cfg, err := config.GetConfig(depl)
	utils.Bail("Could not get config", err)
	s, err := suggest.Analyse(cfg, depls)
	utils.Bail("Suggestion failed", err)
	fmt.Printf("Executing %d start requests...\n", len(s.Starts))
	max_tries := 5
	for _, start := range s.Starts {
		for i := 0; i < max_tries; i++ {
			ctx := authremote.Context()
			fmt.Printf("Deploying %s...\n", start.String())
			d := start.DeployRequest()
			_, err = depl.DeployAppOnTarget(ctx, d)
			if err == nil || *continue_on_error {
				break
			}
			fmt.Printf("Attempt %d of %d - Failed to deploy %v: %s\n", (i + 1), max_tries, start, err)
			time.Sleep(time.Duration(5) * time.Second)
		}
	}
	fmt.Printf("Executing %d stop requests...\n", len(s.Stops))
	for _, stop := range s.Stops {
		for i := 0; i < 5; i++ {
			d := stop.UndeployRequest()
			ctx := authremote.Context()
			fmt.Printf("Undeploying %s...\n", stop.String())
			_, err = depl.UndeployAppOnTarget(ctx, d)
			if err == nil {
				break
			}
			fmt.Printf("Failed to apply %v: %s\n", stop, err)
			time.Sleep(time.Duration(5) * time.Second)
		}
	}
	fmt.Println(s.String())
	if len(s.Starts) != 0 || len(s.Stops) != 0 {
		os.Exit(1)
	}

}

func listDeployments() {
	ctx := authremote.Context()
	depls, err := depl.GetDeploymentsFromCache(ctx, &common.Void{})
	utils.Bail("Failed to get deployments from cache", err)
	fmt.Printf("Current Deployments:\n")
	for _, d := range depls.Deployments {
		fmt.Printf("%10s: ", d.Host)
		for _, group := range d.Apps {
			for _, app := range group.Applications {
				fmt.Printf("  %4d x%d ns=%s group=%s repo=%d binary=%s (deploymentid=%s)\n",
					app.ID,
					app.Instances,
					group.Namespace,
					group.GroupID,
					app.RepositoryID,
					app.Binary,
					app.DeploymentID,
				)
			}
		}
	}
}

func listDeployers() {
	t := &utils.Table{}
	ctx := authremote.Context()
	depls, err := depl.GetKnownAutodeployers(ctx, &common.Void{})
	utils.Bail("Failed to get deployers from cache", err)
	t.AddHeaders("IP", "Groups", "GroupCount")
	for _, ad := range depls.Autodeployers {
		t.AddString(ad.IP)
		t.AddString(strings.Join(ad.Groups, " | "))
		t.AddInt(len(ad.Groups))
		t.NewRow()
	}
	fmt.Println(t.ToPrettyString())

}

func delVersion() error {
	ctx := authremote.Context()
	vers := *del_version
	_, err := depl.DeleteVersion(ctx, &pb.DelVersionRequest{Version: vers})
	utils.Bail("Failed to get deployers from cache", err)
	return nil
}
