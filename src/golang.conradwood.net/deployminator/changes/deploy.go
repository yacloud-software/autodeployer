package changes

import (
	"context"
	"flag"
	"fmt"
	ad "golang.conradwood.net/apis/autodeployer"
	pb "golang.conradwood.net/apis/deployminator"
	dm "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deployminator/config"
	"golang.conradwood.net/deployminator/db"
	"golang.conradwood.net/deployminator/targets"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
	"strings"
	"sync"
	"time"
)

var (
	deploy_live        = flag.Bool("deploy_enable", false, "if true, actually deploy stuff")
	deploy_chan        = make(chan *deploy_req)
	deploy_target_chan = make(chan *deploy_target)
	appcounter         = uint64(0)
	appctrlock         sync.Mutex
)

type deploy_req struct {
	instance *pb.InstanceDef
	dd       *db.FullDD
}
type deploy_target struct {
	appkey uint64
	req    *deploy_req
	target *targets.Target
}

func init() {
	go deploy_worker()
	go deploy_target_worker()
}
func next_app_counter() uint64 {
	appctrlock.Lock()
	if appcounter > 2000000 {
		appcounter = 1
	}
	appcounter++
	res := appcounter
	appctrlock.Unlock()
	return res
}
func submit_deploy(inst *pb.InstanceDef) {
	deploy_chan <- &deploy_req{instance: inst}
}

func deploy_worker() {
	for {
		c := <-deploy_chan
		ctx := authremote.Context()
		dep, err := db.Descriptordb.ByID(ctx, c.instance.DeploymentID.ID)
		if err != nil {
			fmt.Printf("failed to get descriptor: %s\n", err)
			continue
		}
		dd, err := db.FetchFull(ctx, dep)
		if err != nil {
			fmt.Printf("Failed to fetch full: %s\n", err)
			continue
		}
		c.dd = dd
		tl := targets.GetTargets()
		for _, t := range tl.Targets() {
			t.ClearPendingApps()
		}

		mg := c.instance.MachineGroup
		if mg == "" {
			mg = "worker"
		}
		if !config.IgnoreMachineGroups() {
			tl = tl.FilterByMachineGroup(mg)
		}
		if c.instance.InstanceCountIsPerMachine {
			panic("cannot do instancecount yet")
		} else {
			remaining := c.instance.Instances
			for remaining > 0 {
				//	fmt.Printf("Targets: %d\n", len(tl.Targets()))
				ntl := tl.TargetsWithLeastInstances()
				if len(ntl.Targets()) == 0 {
					fmt.Printf("No targets available.\n")
					break
				}
				for _, t := range ntl.Targets() {
					ac := next_app_counter()
					dt := &deploy_target{appkey: ac, target: t, req: c}
					t.AddPendingApp(ac, dd)
					deploy_target_chan <- dt
					remaining--
					if remaining == 0 {
						break
					}
				}
				ntl = tl.TargetsWithLeastInstances()
			}
		}

	}
}

func deploy_target_worker() {
	for {
		c := <-deploy_target_chan
		adl := Lock_Autodeployer(c.target.Host())
		ddl := Lock_DeploymentDescriptor(c.req.instance.DeploymentID)
		err := deploy(c)
		if err != nil {
			fmt.Printf("Error deploying: %s\n", utils.ErrorString(err))
		} else {
			if c.target != nil {
				c.target.RemovePendingApp(c.appkey)
				c.target.Scan()
			}
		}
		adl.Unlock()
		ddl.Unlock()

	}
}
func (dt *deploy_target) GetArguments() ([]*pb.Argument, error) {
	ctx := authremote.Context()
	args, err := db.Argsdb.ByInstanceDef(ctx, dt.req.instance.ID)
	if err != nil {
		return nil, err
	}
	return args, nil
}

func deploy(c *deploy_target) error {
	if !*deploy_live {
		if *debug {
			fmt.Printf("Deploy disabled. see -deploy_enable flag")
		}
		return nil
	}
	fmt.Printf("Deploying on %s\n", c.target)
	ctx := authremote.Context()
	fdd := c.req.dd
	dd := fdd.DeploymentDescriptor
	lm := &dm.Limits{MaxMemory: 3000, Priority: 4}
	url := dd.Application.DownloadURL
	url = strings.ReplaceAll(url, "${BUILDID}", fmt.Sprintf("%d", dd.BuildNumber))
	dargs, err := c.GetArguments()
	if err != nil {
		return err
	}
	var args []string
	for _, a := range dargs {
		args = append(args, a.Argument)
	}
	deplid := fmt.Sprintf("DI-INSTANCEDEF-%d", c.req.instance.ID)
	dr := &ad.DeployRequest{
		DownloadURL:  url,
		Binary:       dd.Application.Binary,
		BuildID:      dd.BuildNumber,
		DeploymentID: deplid,
		Namespace:    "foobar-namespace",
		Groupname:    "foobar-groupname",
		Limits:       lm,
		Args:         args,
		AppReference: &dm.AppReference{
			ID: dd.Application.ID,
			AppDef: &dm.ApplicationDefinition{
				DownloadURL:  url,
				Binary:       dd.Application.Binary,
				BuildID:      dd.BuildNumber,
				DeploymentID: deplid,
				RepositoryID: dd.Application.RepositoryID,
				Limits:       lm,
				Args:         args,
			},
		},
		RepositoryID: dd.Application.RepositoryID,
	}

	d, err := c.target.Autodeployer().Deploy(ctx, dr)
	if err != nil {
		fmt.Printf("Failed to deploy:%s\n", utils.ErrorString(err))
		return err
	}
	id := d.ID
	for {
		ctx := authremote.Context()
		dp, err := get_deployment_by_id(ctx, c.target.Autodeployer(), id)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		if dp == nil {
			return fmt.Errorf("Failed to deploy!!\n")
		}
		if dp.Status == ad.DeploymentStatus_EXECUSER {
			fmt.Printf("Deployed\n")
			return nil
		}
		if dp != nil {
			fmt.Printf("Status: %v\n", dp.Status)
		}
		time.Sleep(1 * time.Second)
	}
}

func get_deployment_by_id(ctx context.Context, ac ad.AutoDeployerClient, id string) (*ad.DeployedApp, error) {
	depls, err := ac.GetDeployments(ctx, &ad.InfoRequest{})
	if err != nil {
		return nil, err
	}
	for _, d := range depls.Apps {
		if d.ID == id {
			return d, nil
		}
	}
	return nil, nil
}
