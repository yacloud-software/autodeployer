package main

// here we work out what we deployed.
// we essentially return the same diff as a filediff
// so we can use the same "diff" function and apply the diffs to deployment

// we get our state by looking for all registered AutoDeployers in the registry
// we then query each one to figure out what they currently have deployed

// we should keep the dependency on other services to a minimum
// afterall this is where we deploy them, so they might not be available
// ATM we need registry, auth and autodeploy

import (
	"context"
	"errors"
	"flag"
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	pb "golang.conradwood.net/apis/deploymonkey"
	rpb "golang.conradwood.net/apis/registry"
	dc "golang.conradwood.net/deploymonkey/common"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/cmdline"
	"golang.conradwood.net/go-easyops/prometheus"
	"golang.conradwood.net/go-easyops/utils"
	"google.golang.org/grpc"
	"strings"
	"time"
)

const (
	DEPLOY_PREFIX = "DM-APPDEF-"
	MIN_RUN_TIME  = 15
	MAX_PROC_MB   = 3000
)

var (
	use_new_style = flag.Int("use_new_deploy_style", 1, "a new experimental deployment style (0=never, 1=for instancesperautodeployer only, 2=always")
	cacheGauge    = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "deploymonkey_precache_status",
			Help: "V=1 UNIT=percent DESC=percentage of pre cache downloads completed",
		},
		[]string{"host"},
	)
	precache                 = flag.Bool("pre_cache", true, "if true, pre cache binaries before attempting deployment")
	max_download_no_progress = flag.Int("max_download_no_progress", 120, "maximum time to allow downloads to hang without progress before failing")
	set_limits               = flag.Bool("set_limits", true, "tell autodeployer to set runtime limits for processes")
	asyncMaker               = make(chan miso, 100)
	incAppId                 = flag.Bool("inc_appid_in_deplid", true, "if true will include the appid in the deploymentid to make it unique. backwards incompatible break, thus disabled by default")
)

type miso struct {
	version int
	ads     []*pb.ApplicationDefinition
}

func init() {
	prometheus.MustRegister(cacheGauge)
	go MakeItSoLoop()
}

// this is the most simplest, but definitely not The Right Thing to do
// how it *should* work:
// * work out what is currently deployed
// * work out a difference
// * fire up all the additional ones required (in parallel)
// * retry failured deployments on different servers
// * if any deployment failed: clear the "new ones" again and abort
// * if all succeeded:
// * clear those which are no longer needed (e.g. old ones in a lower version)
func MakeItSo(ads []*pb.ApplicationDefinition, version int) error {
	// brief sanity check for common stuff...
	for _, ad := range ads {
		if ad.BuildID == 0 {
			return fmt.Errorf("Refusing to deploy application %s with buildid #0", ad.Binary)
		}
	}
	fmt.Printf("Request to upgrade \n")
	m := miso{ads: ads, version: version}
	asyncMaker <- m
	return nil
}

func MakeItSoLoop() {
	t := time.Duration(3) * time.Second
	time.Sleep(t)
	for {
		m := <-asyncMaker
		MakeItSoAsync(m)
	}
}

func MakeItSoAsync(m miso) error {
	//group := m.group

	// split appdefs into "old" and "new" codepath
	var new_path []*pb.ApplicationDefinition
	var old_path []*pb.ApplicationDefinition
	for _, appdef := range m.ads {
		if *use_new_style == 1 {
			if appdef.Instances != 0 && appdef.InstancesMeansPerAutodeployer {
				new_path = append(new_path, appdef)
				continue
			}
		}
		if *use_new_style == 2 {
			new_path = append(new_path, appdef)
			continue
		}
		old_path = append(old_path, appdef)
	}

	err := makeitso_new(new_path)
	if err != nil {
		return err
	}
	if len(old_path) == 0 {
		fmt.Printf("no oldstyle stuff to do for this deployment.\n")
		return nil
	}
	//	fmt.Printf("Applying group %v, version %d\n", group, m.version)

	sas, err := GetDeployers()
	if err != nil {
		return err
	}

	// check all apps are in the same group
	var group *pb.AppGroup
	ctx := context.Background()
	for _, app := range old_path {
		g, err := groupHandler.GetGroupForApp(ctx, app)
		if err != nil {
			return err
		}
		if group == nil {
			group = g
		} else {
			if group.ID != g.ID {
				s := fmt.Sprintf("ERROR: Different groups to deploy (%d and %d)", group.ID, g.ID)
				fmt.Println(s)
				return fmt.Errorf("%s", s)
			}
		}
	}
	if group == nil {
		return fmt.Errorf("no group. no apps?\n")
	}
	groupid := group.ID
	// deploymentid is "PREFIX-GroupID-BuildID"
	// stop all for groupid
	stopPrefix := fmt.Sprintf("%s-%d-", DEPLOY_PREFIX, groupid)
	trans, err := stop(stopPrefix)
	if err != nil {
		return fmt.Errorf("Failed to stop current instances: %s\n", err)
	}
	// starting stuff
	// also, this should start them up multi-threaded... and bla
	var res_err error
	workeridx := 0
	startupids := make(map[string]*rpb.ServiceAddress) // startupid -> autodeployer
	var user_messages []string
	for _, app := range old_path {
		mgroup := app.Machines
		fsas, err := getDeployersInGroup(mgroup, sas)
		if err != nil {
			fmt.Printf("Could not get deployers for group \"%s\": %s\n", mgroup, err)
		}
		if (fsas == nil) || (len(fsas) == 0) {
			s := fmt.Sprintf("No deployers to deploy on for group \"%s\" (app=%v)", mgroup, old_path)
			fmt.Println(s)
			user_messages = append(user_messages, s)
			cancelStop(trans, user_messages, s)
			return errors.New(s)
		}
		workers := len(fsas)
		fmt.Printf("Got %d hosts to deploy on\n", workers)
		fmt.Printf("Starting %d instances of %d\n", app.Instances, app.RepositoryID)
		instances := 0

		retries := 5
		for uint32(instances) < app.Instances {
			if retries == 0 {
				s := fmt.Sprintf("Wanted to deploy %d instances of %s, but only deployed %d on %s", app.Instances, AppToString(app), instances, app.Machines)
				fmt.Println(s)
				user_messages = append(user_messages, s)
				res_err = errors.New(s)
				break
			}
			workeridx++
			if workeridx >= workers {
				workeridx = 0
			}
			autodeployer := fsas[workeridx]
			startupid, msg, terr := deployOn(autodeployer, app)
			if terr == nil {
				startupids[startupid] = autodeployer
				instances++
				retries = 5
				continue
			}
			s := msg + "\n" + fmt.Sprintf("ERROR: %s", terr)
			user_messages = append(user_messages, s)
			time.Sleep(1)
			// deadline expired? reset context
			retries--
			fmt.Printf("failed to deploy an instance: %s (retries=%d)\n", terr, retries)
		}
	}
	if res_err != nil {
		cancelStop(trans, user_messages, fmt.Sprintf("%s", res_err))
	} else {
		if *debug { // NOT A DEBUG IF CLAUSE
			fmt.Printf("Deployed %d instances:\n", len(startupids))
			for k, v := range startupids {
				fmt.Printf("  %s on %s\n", k, v.Host)
			}
		}
		for k, v := range startupids {
			stopperRunningCondition(trans, k, v, MIN_RUN_TIME)
		}
	}
	return err
}

func replaceVars(text string, vars map[string]string) string {
	s := text
	for k, v := range vars {
		s = strings.Replace(s, fmt.Sprintf("${%s}", k), v, -1)
	}
	return s
}

// deploys an instance
// returns deploymentid,usermessage,error
func deployOn(sa *rpb.ServiceAddress, app *pb.ApplicationDefinition) (string, string, error) {
	ctx := authremote.Context()
	gr, err := groupHandler.GetGroupForApp(ctx, app)
	if err != nil {
		return "", "", err
	}
	groupid := gr.ID
	fmt.Printf("Deploying app on host %s:\n", sa.Host)
	dc.PrintApp(app)
	conn, err := DialService(sa)
	if err != nil {
		fmt.Printf("Failed to connect to service %v\n", sa)
		return "", "failed to connect to server", err
	}
	defer conn.Close()

	vars := make(map[string]string)
	vars["BUILDID"] = fmt.Sprintf("%d", app.BuildID)
	vars["REPOSITORYID"] = fmt.Sprintf("%d", app.RepositoryID)
	deplid := fmt.Sprintf("%s-%d-%d", DEPLOY_PREFIX, groupid, app.BuildID)
	if *incAppId {
		deplid = fmt.Sprintf("%s-%d-%d-%d", DEPLOY_PREFIX, groupid, app.BuildID, app.ID)
	}
	adc := ad.NewAutoDeployerClient(conn)
	dr := dc.CreateDeployRequest(nil, app)

	dr.DownloadURL = replaceVars(app.DownloadURL, vars)
	dr.DeploymentID = deplid
	dr.AppReference = &pb.AppReference{ID: 34534534, AppDef: app}

	if *set_limits {
		if app.Limits == nil {
			// given that all config files & db go through an automatic common.AppLimit() to fix up empty AppLimits, this should never happen.
			panic("Program error. shutting down instead of breaking deployments. fix me.")
		}
	}
	if *precache {
		err = waitForCacheStatus(adc, dr, sa.Host)
		if err != nil {
			fmt.Printf("Failed to check cache status. presumed to be old autodeployer. continuing with deploy (%s)\n", err)
		}
	}
	fmt.Printf("Sending deploy request to %s...\n", sa)
	ctx = authremote.Context()
	ad_lock := lockAutodeployerHost(sa.Host)
	defer ad_lock.Unlock()

	dres, err := adc.Deploy(ctx, dr)
	if err != nil {
		s := fmt.Sprintf("failed to deploy #%d(%d) on %v: %s\n", app.BuildID, app.RepositoryID, adc, err)
		fmt.Print(s)
		return "", s, err
	}
	if !dres.Success {
		s := fmt.Sprintf("failed to startup app %v (%s)\n", app, dres.Message)
		fmt.Println(s)
		return "", s, errors.New(s)
	}
	s := fmt.Sprintf("Successfully deployed %v on %s as %s [%s]\n", AppToString(app), sa.Host, deplid, dres.ID)
	fmt.Print(s)
	return dres.ID, s, nil
}

/*
	check if a deploy is in cache, if not start download for it

timeout if no progress was reported for a while
allow for some queries to autodeployer to fail, but not permanently
*/
func waitForCacheStatus(adc ad.AutoDeployerClient, dr *ad.DeployRequest, host string) error {
	ad_lock := lockAutodeployerHost(host)
	defer ad_lock.Unlock()
	fmt.Printf("Checking cache status of %s on %s...\n", dr.Binary, host)
	lastChanged := time.Now()
	lastBytes := uint64(0)
	ureq := &ad.URLRequest{URL: dr.DownloadURL}
	query_succeeded := false
	var lastResponse *ad.URLResponse
	lastResponseReceived := time.Now()
	for {
		ctx := authremote.Context()
		ur, err := adc.CacheURL(ctx, ureq)
		if err != nil {
			if !query_succeeded {
				return err
			}
			if time.Since(lastResponseReceived) > time.Duration(*max_download_no_progress)*time.Second {
				setPreCacheGauge(host, 0, 0)
				return fmt.Errorf("no query responsese for %0.1f seconds on %s (host %s) (last error: %s)", time.Since(lastChanged).Seconds(), ureq.URL, host, utils.ErrorString(err))
			}
			time.Sleep(time.Duration(2) * time.Second)
			continue
		}
		lastResponseReceived = time.Now()
		lastResponse = ur

		query_succeeded = true
		if lastResponse.BytesDownloaded == lastResponse.TotalBytes {
			fmt.Printf("cache status of %s on %s completed (%d of %d bytes)\n", dr.Binary, host, lastResponse.BytesDownloaded, lastResponse.TotalBytes)
			setPreCacheGauge(host, lastResponse.TotalBytes, lastResponse.BytesDownloaded)
			return nil
		}
		if lastResponse.BytesDownloaded != lastBytes {
			lastChanged = time.Now()
		}
		if time.Since(lastChanged) > time.Duration(*max_download_no_progress)*time.Second {
			setPreCacheGauge(host, 0, 0)
			return fmt.Errorf("no download update for %0.1f seconds on %s (host %s)", time.Since(lastChanged).Seconds(), ureq.URL, host)
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
}
func setPreCacheGauge(host string, total, done uint64) {
	l := prometheus.Labels{"host": host}
	perc := float64(0)
	if total != 0 {
		perc = float64(done) / float64(total) * 100
	}
	cacheGauge.With(l).Set(perc)
}

func getDeployments(adc ad.AutoDeployerClient, sa *rpb.ServiceAddress, deplid string) ([]*ad.DeployedApp, error) {
	//	var res []*pb.ApplicationDefinition
	var res []*ad.DeployedApp
	ctx := authremote.Context()
	info, err := adc.GetDeployments(ctx, dc.CreateInfoRequest())
	if err != nil {
		fmt.Printf("Failed to query service %v: %s\n", sa, err)
		return nil, err
	}
	for _, app := range info.Apps {
		dr := app.Deployment
		if !strings.HasPrefix(dr.DeploymentID, deplid) {
			continue
		}
		res = append(res, app)
	}
	return res, nil
}

// given a name will only return deployers in that group name
// if name == "" it will be assumed to be "worker"
func getDeployersInGroup(name string, all []*rpb.ServiceAddress) ([]*rpb.ServiceAddress, error) {
	var res []*rpb.ServiceAddress

	if name == "" {
		name = "worker"
	}
	for _, sa := range all {
		conn, err := DialService(sa)
		if err != nil {
			fmt.Printf("Failed to connect to service %v\n", sa)
			continue
		}
		ctx := authremote.Context()
		adc := ad.NewAutoDeployerClient(conn)
		req := &ad.MachineInfoRequest{}
		mir, err := adc.GetMachineInfo(ctx, req)
		if err != nil {
			conn.Close()
			fmt.Printf("Failed to get machine info on %v\n", sa)
			continue
		}
		conn.Close()
		mg := mir.MachineGroup
		if len(mg) == 0 {
			mg = []string{"worker"}
		}
		fmt.Printf("Autodeployer on %s is in group %s (requested: %s)\n", sa.Host, mg, name)

		if ContainsGroup(mg, name) {
			res = append(res, sa)
		}
	}

	return res, nil
}

// get all registered deployers and their RPC address
func GetDeployers() ([]*rpb.ServiceAddress, error) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(cmdline.GetRegistryAddress(), opts...)
	if err != nil {
		fmt.Printf("Error dialling registry @ %s\n", cmdline.GetRegistryAddress())
		return nil, err
	}
	defer conn.Close()
	rcl := rpb.NewRegistryClient(conn)
	ctx := authremote.Context()
	lr := rpb.V2ListRequest{}
	lr.NameMatch = "autodeployer.AutoDeployer"
	resp, err := rcl.ListRegistrations(ctx, &lr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error getting services: %s", err))
	}
	var res []*rpb.ServiceAddress
	for _, r := range resp.Registrations {
		sa := &rpb.ServiceAddress{
			Host:        r.Target.IP,
			Port:        int32(r.Target.Port),
			ApiType:     r.Target.ApiType,
			RoutingInfo: r.Target.RoutingInfo,
		}
		res = append(res, sa)
		//fmt.Printf("Counting autodeployer at %s\n", r.Target.IP)
	}
	return res, nil
}

type StopRequest struct {
	Binary       string
	RepositoryID uint64
}

// return list of hosts
func stopSingleApp(stop *StopRequest) ([]string, error) {
	fmt.Println("get deployers")
	var res []string
	sas, err := GetDeployers()
	if err != nil {
		return nil, fmt.Errorf("Unable to get deployers: %s", err)
	}
	//stopPrefix := fmt.Sprintf("%s-%d-", DEPLOY_PREFIX, group.id)
	for _, sa := range sas {
		conn, err := DialService(sa)
		if err != nil {
			fmt.Printf("Failed to connect to service %s (%s)\n", sa.Host, err)
			continue
		}
		adc := ad.NewAutoDeployerClient(conn)

		apps, err := getDeployments(adc, sa, "")

		//	fmt.Printf("apps %v\n", apps)

		if err != nil {
			conn.Close()
			return nil, errors.New(fmt.Sprintf("Unable to get deployments from %v: %s", sa, err))
		}
		for _, ap := range apps {
			a := ap.Deployment
			if (a.Binary != stop.Binary) || (a.RepositoryID != stop.RepositoryID) {
				continue
			}
			an := fmt.Sprintf("%s/%s/%d/%s on %s (%s) (%s)", a.Namespace, a.Groupname, a.RepositoryID, a.Binary, sa.Host, a.DeploymentID, ap.ID)
			res = append(res, sa.Host)
			fmt.Printf("Undeploying: %s\n", an)

			ud := ad.UndeployRequest{ID: ap.ID}
			_, err = adc.Undeploy(authremote.Context(), &ud)
			if err != nil {
				fmt.Printf("Failed to shutdown %s (%s)\n", an, err)
			}

		}
		conn.Close()
	}
	return res, nil
}

// if one needs to, one can still connect explicitly to a service
// this call DOES NOT CLOSE THE CONNECTION
func DialService(sa *rpb.ServiceAddress) (*grpc.ClientConn, error) {
	serverAddr := fmt.Sprintf("%s:%d", sa.Host, sa.Port)
	//fmt.Printf("Dialling service at \"%s\"\n", serverAddr)

	creds := client.GetClientCreds()
	cc, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(creds))

	//	opts = []grpc.DialOption{grpc.WithInsecure()}
	// cc, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		fmt.Printf("Error dialling servicename @ %s\n", serverAddr)
		return nil, err
	}
	//defer cc.Close()

	return cc, nil
}

// true if this group is in the deployer
func ContainsGroup(a []string, b string) bool {
	for _, x := range a {
		if x == b {
			return true
		}
	}
	return false
}
