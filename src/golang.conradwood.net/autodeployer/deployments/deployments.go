package deployments

import (
	"fmt"
	pb "golang.conradwood.net/apis/autodeployer"
	dm "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/go-easyops/logger"
	"io"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"
)

const (
	MAX_MEM_MB = 3000
)

var (
	deployed []*Deployed
)

// information about a currently deployed application
type Deployed struct {
	// if true, then there is no application currently deployed for this user
	LastUsed       time.Time
	Idle           bool
	StartupMsg     string
	Status         pb.DeploymentStatus
	Ports          []int
	User           *user.User
	Cmd            *exec.Cmd
	ExitCode       error
	Args           []string
	WorkingDir     string
	Stdout         io.Reader
	Started        time.Time
	Finished       time.Time
	LastLine       string
	Deploymentpath string
	Logger         *logger.AsyncLogQueue
	Pid            uint64 // the pid of starter.go
	Cgroup         int
	DeployRequest  *pb.DeployRequest
	ResolvedArgs   []string
}

func (d *Deployed) DeployedApp() *pb.DeployedApp {
	da := &pb.DeployedApp{
		ID:             d.StartupMsg,
		Deployment:     d.DeployInfo(),
		Ports:          d.PortsUint32(),
		Status:         d.Status,
		RuntimeSeconds: uint64(time.Now().Unix() - d.Started.Unix()),
		DeployRequest:  d.DeployRequest,
	}
	return da
}
func (d *Deployed) PortsUint32() []uint32 {
	var res []uint32
	for _, p := range d.Ports {
		res = append(res, uint32(p))
	}
	return res
}
func (d *Deployed) DeployInfo() *pb.DeployInfo {
	dr := &pb.DeployInfo{
		Status:           d.Status,
		DownloadURL:      d.DeployRequest.DownloadURL,
		DownloadUser:     d.DeployRequest.DownloadUser,
		DownloadPassword: d.DeployRequest.DownloadPassword,
		Binary:           d.DeployRequest.Binary,
		RepositoryID:     d.DeployRequest.RepositoryID,
		BuildID:          d.DeployRequest.BuildID,
		DeploymentID:     d.DeployRequest.DeploymentID,
		Args:             d.DeployRequest.Args,
		AppReference:     d.AppReference(),
		RuntimeSeconds:   uint64(time.Now().Unix() - d.Started.Unix()),
		Ports:            d.PortsUint32(),
		ResolvedArgs:     d.ResolvedArgs,
	}

	return dr
}

func (d *Deployed) Log(format string, a ...interface{}) {
	dn := d.Logger
	if dn == nil {
		fmt.Printf("No logger! (%s)", fmt.Sprintf(format, a...))
		return
	}
	s := format
	if a != nil {
		s = fmt.Sprintf(format, a...)
	}
	dn.LogCommandStdout(s, fmt.Sprintf("%s", d.Status))
}
func (d *Deployed) AppReference() *dm.AppReference {
	return d.DeployRequest.AppReference
}
func (d *Deployed) DeployInfoString() string {
	da := d.AppReference()
	if da == nil {
		return "noappref"
	}
	app := da.AppDef
	if app == nil {
		return "noappdef"
	}
	return fmt.Sprintf("buildid=%d, instances=%d, machines=%s, deplid=%s, repoid=%d", app.BuildID, app.Instances, app.Machines, app.DeploymentID, app.RepositoryID)

}

func (d *Deployed) String() string {
	if d == nil {
		return fmt.Sprintf("[nil Deployed]")
	}
	return fmt.Sprintf("%d-%d (%s) %s", d.DeployRequest.RepositoryID, d.DeployRequest.BuildID, d.StartupMsg, d.Status)
}
func (d *Deployed) GenericString() string {
	return fmt.Sprintf("%s/%s/%d/%s-%d (%s) %s", d.DeployRequest.Namespace, d.DeployRequest.Groupname, d.DeployRequest.RepositoryID, d.DeployRequest.Binary, d.DeployRequest.BuildID, d.StartupMsg, d.Status)
}

func Deployments() []*Deployed {
	return deployed
}
func ActiveDeployments() []*Deployed {
	var res []*Deployed
	for _, d := range deployed {
		if !d.Idle {
			res = append(res, d)
		}
	}
	return res
}

func DownloadingDeployments() []*Deployed {
	var res []*Deployed
	for _, d := range deployed {
		if d.Status == pb.DeploymentStatus_DOWNLOADING {
			res = append(res, d)
		}
	}
	return res
}

func (du *Deployed) GetPortByName(name string) int {
	if !strings.HasPrefix(name, "${PORT") {
		du.Log("Invalid port name: %s", name)
		return 0
	}
	psn := name[6 : len(name)-1]
	pn, err := strconv.Atoi(psn)
	if err != nil {
		du.Log("Could not convert port by name %s to portnumber: %s", name, err)
		return 00
	}
	if pn <= 0 {
		du.Log("Request for port %d (%s) is invalid", pn, name)
		return 0
	}
	pn--
	if pn >= len(du.Ports) {
		du.Log("Port %d not allocated (%d)", pn, len(du.Ports))
		return 0
	}
	po := du.Ports[pn]
	du.Log("Port %s == %d", psn, po)
	return po
}

func Add(d *Deployed) {
	deployed = append(deployed, d)
}

func (d *Deployed) GetExitCode() int {
	if d.ExitCode == nil {
		return 0
	}
	return 10
}

/*
func (d *Deployed) SetLimits(rl *dm.Limits) {
	d.limits = rl
}
*/
func (d *Deployed) Limits() *dm.Limits {
	if d.DeployRequest == nil {
		return nil
	}
	if d.DeployRequest.Limits != nil {
		return d.DeployRequest.Limits
	}
	return nil
}
func (d *Deployed) RepositoryID() uint64 {
	if d.DeployRequest == nil {
		return 0
	}
	return d.DeployRequest.RepositoryID
}
func (d *Deployed) URL() string {
	if d.DeployRequest == nil {
		return ""
	}
	return d.DeployRequest.DownloadURL
}
func (d *Deployed) Namespace() string {
	if d.DeployRequest == nil {
		return ""
	}
	return d.DeployRequest.Namespace
}
func (d *Deployed) Groupname() string {
	if d.DeployRequest == nil {
		return ""
	}
	return d.DeployRequest.Groupname
}
func (d *Deployed) Binary() string {
	if d.DeployRequest == nil {
		return ""
	}
	return d.DeployRequest.Binary
}
func (d *Deployed) DeploymentID() string {
	if d.DeployRequest == nil {
		return ""
	}
	return d.DeployRequest.DeploymentID
}
func (d *Deployed) Build() uint64 {
	if d.DeployRequest == nil {
		return 0
	}

	return d.DeployRequest.BuildID
}
func (d *Deployed) AutoRegistrations() []*dm.AutoRegistration {
	if d.DeployRequest == nil {
		return nil
	}
	return d.DeployRequest.AutoRegistration
}

/*******************************************************
// cgroup stuff
*******************************************************/
func (d *Deployed) GetCgroup() int {
	return d.Cgroup
}
func (d *Deployed) SetCgroup(i int) {
	d.Cgroup = i
}
func (d *Deployed) GetPid() uint64 {
	return d.Pid
}
