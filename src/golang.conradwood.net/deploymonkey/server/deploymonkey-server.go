package main

import (
	"context"
	"database/sql"

	//	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/lib/pq"
	apb "golang.conradwood.net/apis/autodeployer"
	common "golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/deploymonkey"
	rpb "golang.conradwood.net/apis/registry"
	dc "golang.conradwood.net/deploymonkey/common"
	"golang.conradwood.net/deploymonkey/db"
	"golang.conradwood.net/deploymonkey/scheduler"
	"golang.conradwood.net/deploymonkey/useroverride"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/server"
	gesql "golang.conradwood.net/go-easyops/sql"
	"golang.conradwood.net/go-easyops/utils"
	"google.golang.org/grpc"
)

// static variables for flag parser
var (
	debug = flag.Bool("debug", false, "more logging...")
	limit = flag.Int("limit", 100, "max entries to return when querying for lists")
	port  = flag.Int("port", 4999, "The server port")
	/*
		dbhost           = flag.String("dbhost", "localhost", "hostname of the postgres database rdms")
		dbdb             = flag.String("database", "deploymonkey", "database to use for authentication")
		dbuser           = flag.String("dbuser", "root", "username for the database to use for authentication")
		dbpw             = flag.String("dbpw", "pw", "password for the database to use for authentication")
	*/
	applyonly        = flag.Bool("apply_only", false, "if true will apply current config and exit")
	enableScanner    = flag.Bool("enable_scanner", true, "if true will periodically scan all autodeployers")
	testScanner      = flag.Bool("test_scanner", false, "true will only invoke the scanner")
	testmode         = flag.Bool("testmode", false, "sets some stuff to make it more convenient to test")
	reapply_on_start = flag.Bool("reapply", false, "set to true if you want deploymonkey to reset all versions (shuts down all services and restarts them!!")
	applyinterval    = flag.Int("apply_interval", 60, "`seconds` between scans for discrepancies and re-applying them")
	list             = flag.String("list", "", "list this `repository` previous versions")
	dbcon            *gesql.DB
	dbinfo           string
	applyLock        sync.Mutex
	applyChannel     chan *applyingInfo
	appdef_store     *db.DBApplicationDefinition
)

type applyingInfo struct {
	version int
}

type appVersionDef struct {
	appdef *pb.ApplicationDefinition
	gv     *groupVersion
}
type groupVersion struct {
	Version int
	GroupID int
	Created time.Time
}

// callback from the compound initialisation
func st(server *grpc.Server) error {
	s := new(DeployMonkey)
	// Register the handler object
	pb.RegisterDeployMonkeyServer(server, s)
	return nil
}

func main() {
	var err error
	flag.Parse() // parse stuff. see "var" section above
	dbcon, err = gesql.Open()
	utils.Bail("failed to open postgres", err)
	db.DefaultDBContainerDef().SaveWithID(context.Background(), &pb.ContainerDef{ID: 0})
	appdef_store = db.DefaultDBApplicationDefinition()
	db.DefaultDBAppGroup()
	//	db.DefaultDBGroupVersion()
	utils.Bail("failed to start group2 handler", start_group2_handler())
	if *testScanner {
		ScanAutodeployersTest()
		os.Exit(0)
	}

	applyChannel = make(chan *applyingInfo, 10)
	go func() {
		for {
			ap := <-applyChannel
			ctx := authremote.Context()
			err := applyVersionWithInfo(ctx, ap)
			if err != nil {
				fmt.Printf("Failed to apply %d: %s\n", ap.version, err)
			}
		}
	}()

	if !*testmode {
		applyTimer := time.NewTimer(time.Second * time.Duration(*applyinterval))
		go func() {
			<-applyTimer.C
			applyAllVersions(context.Background(), true)
		}()
	}
	if *enableScanner {
		go func() {
			time.Sleep(time.Duration(3) * time.Second)
			StartScanner()
		}()
	}
	go startVersionGauge()
	sd := server.NewServerDef()
	sd.SetPort(*port)
	sd.SetRegister(st)
	err = server.ServerStartup(sd)
	if err != nil {
		fmt.Printf("failed to start server: %s\n", err)
	}
	fmt.Printf("Done\n")
	return
}

/**********************************
* catch-all fix up
***********************************/
// this gets all groups, all current versions
// and makes the deployment match
// if pendingonly is true, will check for mismatched current != pending versions
// and only apply those
func applyAllVersions(ctx context.Context, pendingonly bool) error {
	var dv int
	if pendingonly {
		fmt.Printf("(Re)applying all pending versions...\n")
	} else {
		fmt.Printf("Reapplying all current versions...\n")
	}
	var err error
	var rows *sql.Rows
	if pendingonly {
		rows, err = dbcon.QueryContext(ctx, "applypendingversions", "SELECT pendingversion from appgroup where deployedversion != pendingversion")
	} else {
		rows, err = dbcon.QueryContext(ctx, "applyallversions", "SELECT deployedversion from appgroup ")
	}
	if err != nil {
		fmt.Printf("Failed to get deployedversions: %s\n", err)
		return err
	}
	for rows.Next() {
		err := rows.Scan(&dv)
		if err != nil {
			fmt.Printf("Failed to get deployedversion for a group: %s\n", err)
			return err
		}
		applyVersion(dv)
	}
	fmt.Printf("Reapply complete\n")
	return nil
}

/**********************************
* implementing the postgres functions here:
***********************************/

func saveApp(app *pb.ApplicationDefinition) (uint64, error) {
	err := dc.CheckAppComplete(app)
	if err != nil {
		return 0, err
	}
	if app.Container != nil {
		_, err = db.DefaultDBContainerDef().Save(TEMPCONTEXT(), app.Container)
		if err != nil {
			return 0, err
		}
	} else {
		app.Container = &pb.ContainerDef{ID: 0}
	}
	app.Created = uint32(time.Now().Unix())
	id, err := appdef_store.Save(TEMPCONTEXT(), app)
	if err != nil {
		return 0, errors.Errorf("Failed to insert application: %s", err)
	}

	for _, arg := range app.Args {
		_, err = dbcon.ExecContext(TEMPCONTEXT(), "insertargs", "INSERT INTO args (argument,app_id) values ($1,$2)", arg, id)
		if err != nil {
			return 0, errors.Errorf("Failed to insert tag for app %d: %s", id, err)
		}
	}
	for _, ar := range app.AutoRegs {
		_, err = dbcon.ExecContext(TEMPCONTEXT(), "foo", "INSERT INTO autoreg (portdef,servicename,apitypes,app_id) values ($1,$2,$3,$4)", ar.Portdef, ar.ServiceName, ar.ApiTypes, id)
		if err != nil {
			return 0, errors.Errorf("Failed to insert autoreg for app %d: %s", id, err)
		}
	}
	_, err = dbcon.ExecContext(TEMPCONTEXT(), "foo", "INSERT into applimits (app_id,maxmemory,priority) values ($1,$2,$3)", id, app.Limits.MaxMemory, app.Limits.Priority)
	if err != nil {
		return 0, errors.Errorf("Failed to insert limits for app %d: %s", id, err)
	}
	return id, nil
}

// given a group version will load all its apps into objects
func getGroupLatestVersion(ctx context.Context, namespace string, groupname string) (int, error) {
	rows, err := dbcon.QueryContext(ctx, "getgrouplatestversion", "SELECT MAX(group_version.id) as maxid from group_version,appgroup where group_id = appgroup.id and appgroup.namespace = $1 and appgroup.groupname = $2", namespace, groupname)
	if err != nil {
		fmt.Printf("Failed to get latest version for (%s,%s):%s\n", namespace, groupname, err)
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var maxid int
		err = rows.Scan(&maxid)
		if err != nil {
			fmt.Printf("Failed to scan for latest version for (%s,%s):%s\n", namespace, groupname, err)
			return 0, err
		}
		return maxid, nil
	}
	return 0, nil
}

// given a group version will load all its apps into objects
func loadAppGroupVersion(ctx context.Context, version uint32) ([]*pb.ApplicationDefinition, error) {
	var res []*pb.ApplicationDefinition
	if version == 0 {
		return res, nil
	}

	//if *testmode {
	//fmt.Printf("Loading appgroup version #%d\n", version)
	//}
	rows, err := dbcon.QueryContext(ctx, "loadappgroupversion", "SELECT lnk_app_grp.app_id from lnk_app_grp where lnk_app_grp.group_version_id = $1", version)
	if err != nil {
		fmt.Printf("Failed to get apps for version %d:%s\n", version, err)
		return nil, errors.Errorf("loadAppGroupVersion(): query failed with %s", err)
	}
	var ids []uint64
	for rows.Next() {
		var id uint64
		err = rows.Scan(&id)
		if err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	for _, id := range ids {
		r, err := loadApp(ctx, id)

		if err != nil {
			fmt.Printf("loadAppGroupVersion(): Failed to load app for version %d (got id==%d from lnk_app_grp.app_id) :%s\n", version, id, err)
			return nil, err
		}
		res = append(res, r)
	}

	return res, nil

}

func loadAppByID(ctx context.Context, id uint64) (*pb.ApplicationDefinition, error) {
	return loadApp(ctx, id)
}

// turns a database row into an applicationdefinition object
// optionally suppling an interface to take up additional values
func loadApp(ctx context.Context, id uint64) (*pb.ApplicationDefinition, error) {

	res, err := appdef_store.ByID(ctx, id)
	if err != nil {
		fmt.Printf("loadApp(): APPDEF store error: %s\n", err)
		//		res, err = ConvertOldApp(ctx, id)
		//		if err != nil {
		return nil, errors.Errorf("loading app #%d caused error: %s", id, err)
		//		}
	}
	if res.Container == nil || res.Container.ID == 0 {
		res.Container = nil
	} else {
		res.Container, err = db.DefaultDBContainerDef().ByID(ctx, res.Container.ID)
		if err != nil {
			return nil, err
		}
	}

	args, err := loadAppArgs(res.ID)
	if err != nil {
		return nil, err
	}
	res.Args = args
	//fmt.Println("loadAutoReg")
	regs, err := loadAutoReg(res.ID)
	if err != nil {
		return nil, err
	}
	res.AutoRegs = regs
	res.Limits, err = loadAppLimits(res.ID)
	dc.AppLimits(res)
	return res, nil
}

// given an application id, loads "app limits"
func loadAppLimits(id uint64) (*pb.Limits, error) {
	// add new limits here...
	rows, err := dbcon.QueryContext(TEMPCONTEXT(), "loadapplimits", "SELECT maxmemory,priority from applimits where app_id = $1", id)
	if err != nil {
		fmt.Printf("Failed to get app with id %d:%s\n", id, err)
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	res := &pb.Limits{}
	err = rows.Scan(&res.MaxMemory, &res.Priority)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// given an application, it loads the args from DB
func loadAppArgs(id uint64) ([]string, error) {
	var res []string
	var s string
	rows, err := dbcon.QueryContext(TEMPCONTEXT(), "loadappargs", "SELECT argument from args where app_id = $1", id)
	if err != nil {
		s := fmt.Sprintf("Failed to get tags for app %d:%s\n", id, err)
		return nil, errors.Errorf("%s", s)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&s)
		if err != nil {
			s := fmt.Sprintf("Failed to get tag for app %d:%s\n", id, err)
			return nil, errors.Errorf("%s", s)
		}
		res = append(res, s)
	}
	return res, nil
}

// given an applicationid, it loads the args from DB
func loadAutoReg(id uint64) ([]*pb.AutoRegistration, error) {
	var res []*pb.AutoRegistration
	if *testmode {
		fmt.Printf("Loading auto registration for app_id %d\n", id)
	}
	rows, err := dbcon.QueryContext(TEMPCONTEXT(), "loadautoreg", "SELECT portdef,servicename,apitypes from autoreg where app_id = $1", id)
	if err != nil {
		s := fmt.Sprintf("Failed to get autoregs for app %d:%s\n", id, err)
		return nil, errors.Errorf("%s", s)
	}
	defer rows.Close()
	for rows.Next() {
		ar := &pb.AutoRegistration{}
		err = rows.Scan(&ar.Portdef, &ar.ServiceName, &ar.ApiTypes)
		if err != nil {
			s := fmt.Sprintf("Failed to get autoreg for app %d:%s\n", id, err)
			return nil, errors.Errorf("%s", s)
		}
		res = append(res, ar)
	}
	return res, nil
}

// get group id from version
func getGroupIDFromVersion(v int) (*groupVersion, error) {
	gr := groupVersion{Version: v}
	err := dbcon.QueryRowContext(TEMPCONTEXT(), "foogroupid", "select group_id,created from group_version where id = $1", v).Scan(&gr.GroupID, &gr.Created)
	if err != nil {
		return nil, err
	}
	return &gr, nil
}

// update the deployed version of a group (group referred to by version!)
func updateDeployedVersionNumber(v int) error {
	gid, err := getGroupIDFromVersion(v)
	if err != nil {
		return errors.Errorf("Invalid Group-Version: \"%d\": %s", v, err)
	}
	_, err = dbcon.ExecContext(TEMPCONTEXT(), "fooexec", "update appgroup set deployedversion = $1 where id = $2", v, gid.GroupID)
	if err != nil {
		return errors.Errorf("Unable to update group: %s", err)
	}
	fmt.Printf("Updated deployedversion to %d\n", v)
	return nil
}

func applyVersion(version int) {
	fmt.Printf("Submitting version %d to apply\n", version)
	x := applyingInfo{version: version}
	applyChannel <- &x
}

// given a version of a group checks the workers and fixes it up to match version
func applyVersionWithInfo(ctx context.Context, curApply *applyingInfo) error {
	rand := utils.RandomString(5)
	fmt.Printf("[thread %s] Waiting for applyVersionLock() (held by %v)...\n", rand, curApply)
	applyLock.Lock()
	defer applyLock.Unlock()
	scheduler.DefaultScheduler.DeployRequested()
	fmt.Printf("[thread %s] Applying version %d\n", rand, curApply.version)
	// first step: mark the version as pending
	// so if it fails for some reason, we know what to replay
	gid, err := getGroupIDFromVersion(curApply.version)
	if err != nil {
		return errors.Errorf("Invalid Group-Version: \"%d\": %s", curApply.version, err)
	}
	_, err = dbcon.ExecContext(TEMPCONTEXT(), "fooexec", "update appgroup set pendingversion = $1 where id = $2", curApply.version, gid.GroupID)
	if err != nil {
		return errors.Errorf("Unable to update group: %s", err)
	}
	var ns, gn string
	err = dbcon.QueryRowContext(TEMPCONTEXT(), "rowfoo", "SELECT namespace,groupname from appgroup where id = $1", gid.GroupID).Scan(&ns, &gn)
	if err != nil {
		return errors.Errorf("Unable to get groupnames: %s", err)
	}

	apps, err := loadAppGroupVersion(ctx, uint32(curApply.version))
	if err != nil {
		return errors.Errorf("error loading apps for version %d: %s", curApply.version, err)
	}
	var f_apps []*pb.ApplicationDefinition
	for _, a := range apps {
		useroverride.MarkAsDeployed(a)
		app, err := loadAppByID(ctx, a.ID)
		if err != nil {
			return err
		}
		f_apps = append(f_apps, app)
	}
	fmt.Printf("Makeitso (%d)...\n", curApply.version)
	err = MakeItSo(f_apps, curApply.version)
	if err != nil {
		return errors.Errorf("error applyings apps for version %d: %s", curApply.version, err)
	}
	go NotifyPeopleAboutDeploy(apps, curApply.version)
	return nil
}

/**********************************
* implementing the server functions here:
***********************************/
type DeployMonkey struct{}

// given a Version# -> Take it online ("Make it so")
func (s *DeployMonkey) DeployVersion(ctx context.Context, cr *pb.DeployRequest) (*pb.DeployResponse, error) {
	if cr.VersionID == "" {
		return nil, errors.Errorf("%s", "VersionID required for deployment")
	}
	version, err := strconv.Atoi(cr.VersionID)
	if err != nil {
		return nil, errors.Errorf("Invalid VersionID: \"%s\": %s", cr.VersionID, err)
	}
	applyVersion(version)
	updateDeployedVersionNumber(version)
	r := pb.DeployResponse{}
	return &r, nil
}

// merge source into target
// basically anything set in source shall be copied to target
// returns true if there was something updated
func mergeApp(t, s *pb.ApplicationDefinition) bool {
	res := false
	if (s.DownloadURL != "") && (s.DownloadURL != t.DownloadURL) {
		res = true
		t.DownloadURL = s.DownloadURL
	}
	if (s.DownloadUser != "") && (s.DownloadUser != t.DownloadUser) {
		res = true
		t.DownloadUser = s.DownloadUser
	}
	if (s.DownloadPassword != "") && (s.DownloadPassword != t.DownloadPassword) {
		res = true
		t.DownloadPassword = s.DownloadPassword
	}
	if (len(s.Args) != 0) && (!AreArgsIdentical(s, t)) {
		res = true
		t.Args = s.Args
	}
	if (s.Binary != "") && (s.Binary != t.Binary) {
		res = true
		t.Binary = s.Binary
	}
	if (s.BuildID != 0) && (s.BuildID != t.BuildID) {
		res = true
		t.BuildID = s.BuildID
	}
	if (s.Instances != 0) && (s.Instances != t.Instances) {
		res = true
		t.Instances = s.Instances
	}
	if s.Java != t.Java {
		res = true
		t.Java = s.Java
	}
	return res
}

func getStringsFromDB(sqls string, val string) ([]string, error) {
	var err error
	var rows *sql.Rows
	if val != "" {
		rows, err = dbcon.QueryContext(TEMPCONTEXT(), "getstrfromdb", sqls, val)
	} else {
		rows, err = dbcon.QueryContext(TEMPCONTEXT(), "getstrfromdb", sqls)
	}
	if err != nil {
		fmt.Printf("Failed to query \"%s\": %s\n", sqls, err)
		return nil, err
	}
	var res []string
	var dv string
	for rows.Next() {
		err := rows.Scan(&dv)
		if err != nil {
			fmt.Printf("Failed to get deployedversion for a group: %s\n", err)
			return nil, err
		}
		res = append(res, dv)
	}
	return res, nil
}

func (s *DeployMonkey) ParseConfigFile(ctx context.Context, cr *pb.ParseRequest) (*pb.ParseConfigResponse, error) {
	fd, err := dc.ParseConfig([]byte(cr.Config), cr.RepositoryID)
	if err != nil {
		return nil, err
	}
	res := pb.ParseConfigResponse{}
	res.GroupDef = fd.Groups
	return &res, nil
}

/*
func (s *DeployMonkey) ApplyVersions(ctx context.Context, cr *pb.ApplyRequest) (*pb.EmptyMessage, error) {
	applyAllVersions(!cr.All)
	return &pb.EmptyMessage{}, nil
}
*/

/*
func (s *DeployMonkey) DeleteApplication(ctx context.Context, dar *pb.DeleteApplicationRequest) (*pb.EmptyMessage, error) {

	_, err := dbcon.Exec(TEMPCONTEXT(),"fooexec","DELETE from appdef where id=$1", dar.ID)
	if err != nil {
		msg := fmt.Sprintf("Error attempting to delete app from appdef with ID=%d %s\n", dar.ID, err)
		fmt.Println(msg)
		return nil, errors.Errorf("%s",msg)
	}

	_, err = dbcon.Exec(TEMPCONTEXT(),"fooexec","DELETE from lnk_app_grp where app_id=$1", dar.ID)
	if err != nil {
		msg := fmt.Sprintf("Error attempting to delete app from appdef with ID=%d %s\n", dar.ID, err)
		fmt.Println(msg)
		return nil, errors.Errorf("%s",msg)
	}

	return &pb.EmptyMessage{}, nil
}
*/

func (s *DeployMonkey) UndeployApplication(ctx context.Context, uar *pb.UndeployApplicationRequest) (*pb.UndeployResponse, error) {
	res := &pb.UndeployResponse{}
	fmt.Printf("Request to stop app version #%d\n", uar.ID)
	pbs, err := loadAppGroupVersion(ctx, uint32(uar.ID))
	if err != nil {
		msg := fmt.Sprintf("Error getting group for ID=%d %s\n", uar.ID, err)
		fmt.Println(msg)
		return nil, errors.Errorf("%s", msg)
	}
	for _, pb := range pbs {
		res.App = pb
		useroverride.MarkAsUndeployed(pb)
		fmt.Printf("Undeploying: %d/%s (%d)\n", pb.RepositoryID, pb.Binary, pb.ID)
		sng, err := stopSingleApp(&StopRequest{RepositoryID: pb.RepositoryID, Binary: pb.Binary})
		if err != nil {
			msg := fmt.Sprintf("Error attempting to stop app with ID=%d %s\n", uar.ID, err)
			fmt.Println(msg)
			return nil, errors.Errorf("%s", msg)
		}
		for _, x := range sng {
			res.Host = append(res.Host, x)
		}
	}
	return res, nil
}

/*
	func (s *DeployMonkey) SetMachineStatus(ctx context.Context, req *pb.SetMachineStatusRequest) (*pb.EmptyMessage, error) {
		if req.Status == nil {
			return nil, errors.Errorf("No status specified (ip=%s)", req.Ip)
		}
		err := errors.Errorf("SetMachineStatus not implemented")
		if err != nil {
			return nil, err
		}
		return &pb.EmptyMessage{}, nil
	}
*/
func (s *DeployMonkey) GetDeploymentsFromCache(ctx context.Context, req *common.Void) (*pb.DeploymentList, error) {
	return INT_GetDeploymentsFromCache(ctx)
}
func INT_GetDeploymentsFromCache(ctx context.Context) (*pb.DeploymentList, error) {
	if !*enableScanner {
		return nil, errors.Errorf("Scanner not enabled")
	}
	dls := GetAllDeployersFromCache()
	res := &pb.DeploymentList{}
	for _, dl := range dls {
		for _, ad := range dl.Apps {
			d := &pb.Deployment{Host: dl.IP}
			g, err := convertDeployedToGroupDef(ctx, ad)
			if err != nil {
				s := "unknown"
				if ad.Deployment != nil {
					add := ad.Deployment
					s = fmt.Sprintf("%s/%s/%d", add.Binary, add.Namespace, add.RepositoryID)
				}
				fmt.Printf("Our cache is dodgy. Converting deployed app (%s) on %s into our own, we got: %s\n", s, dl.IP, err)
				continue
			}
			d.Apps = append(d.Apps, g)
			res.Deployments = append(res.Deployments, d)
		}
	}
	if *debug {
		fmt.Printf("Deployments in cache: %d\n", len(res.Deployments))
		for _, d := range res.Deployments {
			fmt.Printf("   Host: %s\n", d.Host)
			for _, g := range d.Apps {
				for _, app := range g.Applications {
					fmt.Printf("      %s\n", app.Binary)
				}
			}
		}
	}
	return res, nil
}

func convertDeployedToGroupDef(ctx context.Context, app *apb.DeployedApp) (*pb.GroupDefinitionRequest, error) {
	d := app.Deployment
	if d.DeploymentID == "" {
		return nil, errors.Errorf("missing deploymentid unable to convert app %s to groupdef", d.Binary)
	}
	groupid, _, appdefid := DecodeDeploymentID(d.DeploymentID)
	group, err := groupHandler.GroupByID(ctx, uint64(groupid))
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.Errorf("No group with id %d\n", groupid)
	}

	res := &pb.GroupDefinitionRequest{}
	//res := group.groupDef
	var appdef *pb.ApplicationDefinition
	if appdefid > 0 {
		appdef, err = loadAppByID(ctx, uint64(appdefid))
		if err != nil {
			return nil, err
		}
		if appdef.DeploymentID == "" {
			if d.DeploymentID == "" {
				fmt.Printf("WARNING (%#v) has no deploymentid\n", d)
			}
			appdef.DeploymentID = d.DeploymentID
		}
	} else {
		i := true
		if i {
			panic("appdefid is 0")
		}
		// this shouldn't need to be called,
		// once we got all deployed apps with the ID we'll load all of them
		// from database instead (see above)
		appdef = &pb.ApplicationDefinition{DownloadURL: d.DownloadURL,
			DownloadUser:     d.DownloadUser,
			DownloadPassword: d.DownloadPassword,
			Binary:           d.Binary,
			Args:             d.Args,
			RepositoryID:     d.RepositoryID,
			BuildID:          d.BuildID,
			Instances:        0,
			DeploymentID:     d.DeploymentID,
			AutoRegs:         d.AutoRegistration,
			DeployType:       d.DeployType,
			ID:               uint64(appdefid),
			Critical:         false,
		}
	}
	dc.AppLimits(appdef)
	res.Applications = append(res.Applications, appdef)
	return res, nil
}

func (depl *DeployMonkey) DeployAppOnTarget(ctx context.Context, dr *pb.DeployAppRequest) (*common.Void, error) {
	app, err := loadAppByID(ctx, dr.AppID)
	if err != nil {
		return nil, errors.Errorf("Failed to load app #%d: %s", dr.AppID, err)
	}

	sa := &rpb.ServiceAddress{Host: dr.Host, Port: 4000}
	fmt.Printf("Deploying App %v on %s\n", app, dr.Host)
	_, msg, err := deployOn(sa, app)
	if err != nil {
		fmt.Println(msg)
		return nil, err
	}
	sc := &scanner{}
	go sc.ScanAutodeployer(sa) // rescan to keep information somewhat uptodate
	return &common.Void{}, err
}
func (depl *DeployMonkey) UndeployAppOnTarget(ctx context.Context, dr *pb.UndeployAppRequest) (*common.Void, error) {
	match := dr.DeploymentID
	fmt.Printf("Undeploying app \"%s\" on host %s\n", match, dr.Host)
	if match == "" {
		return nil, errors.Errorf("No app to undeploy specified")
	}
	conn, err := DialService(&rpb.ServiceAddress{Host: dr.Host, Port: 4000})
	if err != nil {
		fmt.Printf("Failed to connect to host %v\n", dr.Host)
		return nil, err
	}
	defer conn.Close()
	ir, err := getDeploymentsOnHost(ctx, conn)
	if err != nil {
		return nil, errors.Errorf("Unable to get deployments on host %s: %s\n", dr.Host, err)
	}
	fmt.Printf("to be stopped ID: \"%s\"\n", match)
	var app *apb.DeployedApp
	var app2 *apb.DeployedApp
	for _, d := range ir.Apps {
		fmt.Printf("Deployedapp: %#v\n", d)
		if d.ID == match {
			app = d
			break
		}
		if d.Deployment.DeploymentID == match {
			app2 = d
		}

	}
	if app == nil {
		app = app2
	}
	if app == nil {
		return nil, errors.Errorf("Found no app \"%s\" on host %s\n", match, dr.Host)
	}
	fmt.Printf("host %s: Undeploying app ID \"%s\"\n", dr.Host, app.ID)

	adc := apb.NewAutoDeployerClient(conn)
	_, err = adc.Undeploy(ctx, &apb.UndeployRequest{ID: fmt.Sprintf("%s", app.ID), Block: true})
	if err != nil {
		fmt.Printf("Failed to undeploy: %s\n", err)
		return nil, err
	}

	return &common.Void{}, nil
}
func TEMPCONTEXT() context.Context {
	return context.Background()
}

func (depl *DeployMonkey) GetKnownAutodeployers(ctx context.Context, dr *common.Void) (*pb.AutodeployerList, error) {
	res := &pb.AutodeployerList{}
	for _, sad := range GetAllDeployersFromCache() {
		ad := &pb.AutodeployerInfo{
			IP:        sad.IP,
			Port:      uint32(sad.Port),
			Groups:    sad.Group,
			Broken:    sad.Broken,
			Available: sad.Available,
			LastSeen:  uint32(sad.LastSeen.Unix()),
		}
		res.Autodeployers = append(res.Autodeployers, ad)
	}
	return res, nil
}
