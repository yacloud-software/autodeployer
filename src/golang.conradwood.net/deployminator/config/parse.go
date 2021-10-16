package config

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var (
	ignore_machine_group = flag.Bool("ignore_machine_group", false, "if true, ignore machine groups and consider them all matching")
	parse_strict         = flag.Bool("parse_strict", false, "should be true, but during transition from repository to repositoryid it cannot")
)

type FileDef struct {
	Namespace string
	Groups    []*pb.GroupDefinitionRequest
}

func IgnoreMachineGroups() bool {
	return *ignore_machine_group
}
func PrintSample() {
	foo := &FileDef{Namespace: "example",
		Groups: []*pb.GroupDefinitionRequest{
			&pb.GroupDefinitionRequest{
				Namespace: "example",
				GroupID:   "foogroupid",
				Applications: []*pb.ApplicationDefinition{
					&pb.ApplicationDefinition{
						Args:   []string{"arg1", "arg2"},
						Limits: &pb.Limits{MaxMemory: 3000},
					},
				},
			},
		},
	}
	x, err := yaml.Marshal(foo)
	if err != nil {
		fmt.Printf("Failed to marshal: %s\n", err)
		return
	}
	fmt.Println(string(x))
}

func ParseFile(fname string, repoid uint64) (*FileDef, error) {
	fmt.Printf("Parsing %s\n", fname)
	fb, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Printf("Failed to read file %s: %s\n", fname, err)
		return nil, err
	}
	return Parse(fb, repoid)
}
func Parse(fb []byte, repoid uint64) (*FileDef, error) {
	res, err := ParseConfig(fb, repoid)
	if err != nil {
		fmt.Printf("Failed to parse file: %s\n", err)
		return nil, err
	}
	fmt.Printf("Found %d groups\n", len(res.Groups))
	fmt.Printf("Namespace: %s\n", res.Namespace)
	for _, x := range res.Groups {
		PrintGroup(x)
	}
	return res, nil
}
func ParseConfig(config []byte, repoid uint64) (*FileDef, error) {
	if repoid == 0 {
		return nil, fmt.Errorf("missing parameter repositoryid (0 is not valid)")
	}
	gd := FileDef{}
	var err error
	if *parse_strict {
		err = yaml.UnmarshalStrict(config, &gd)
	} else {
		err = yaml.Unmarshal(config, &gd)
	}
	if err != nil {
		fmt.Printf("Failed to parse yaml: %s\n", err)
		return nil, err
	}
	// apply namespace & repoid throughout
	for _, x := range gd.Groups {
		if x.Namespace == "" {
			x.Namespace = gd.Namespace
		}
	}
	for _, g := range gd.Groups {
		for _, app := range g.Applications {
			if app.RepositoryID == 0 {
				app.RepositoryID = repoid
			}
			AppLimits(app)
			err = CheckAppComplete(app)
			if err != nil {
				return nil, err
			}
		}
	}
	return &gd, nil
}
func AppLimits(app *pb.ApplicationDefinition) *pb.Limits {
	if app.Limits != nil {
		return app.Limits
	}
	// the default applied to all apps w/o limits configured
	app.Limits = &pb.Limits{
		MaxMemory: 3000,
	}
	return app.Limits
}
func CheckAppComplete(app *pb.ApplicationDefinition) error {
	if app.RepositoryID == 0 {
		return fmt.Errorf("Missing RepositoryID")
	}

	if (app.Critical) && (!app.AlwaysOn) {
		return fmt.Errorf("Invalid combination of flags: Cannot have critical app which is not always on")
	}
	if (app.DeployType == "staticfiles") && (app.StaticTargetDir == "") {
		return fmt.Errorf("A staticfiles package REQUIRES StaticTargetDir and it is missing")
	}
	s := fmt.Sprintf("%d-%s", app.RepositoryID, app.DeploymentID)
	if app.DownloadURL == "" {
		return errors.New(fmt.Sprintf("%s is invalid: Missing DownloadURL", s))
	}
	return nil
}

func PrintGroup(x *pb.GroupDefinitionRequest) {
	fmt.Printf("  Group: %s with %d applications\n", x.GroupID, len(x.Applications))
	fmt.Printf("        Namespace  : %s\n", x.Namespace)
	for _, a := range x.Applications {
		fmt.Printf("%s", PrintApp(a))
	}
}
func PrintApp(a *pb.ApplicationDefinition) string {
	var b bytes.Buffer

	fmt.Printf("        Application: (AlwaysOn: %v, Critical %v)\n", a.AlwaysOn, a.Critical)
	fmt.Printf("             Repo  : %d\n", a.RepositoryID)
	fmt.Printf("             Binary: %s\n", a.Binary)
	fmt.Printf("            BuildID: %d\n", a.BuildID)
	fmt.Printf("           Machines: %s\n", a.Machines)
	fmt.Printf("               Type: %s\n", a.DeployType)
	fmt.Printf("             Public: %v\n", a.Public)
	b.WriteString(fmt.Sprintf("           %d Args: ", len(a.Args)))
	for _, arg := range a.Args {
		b.WriteString(fmt.Sprintf("%s ", arg))
	}
	b.WriteString("%\n")
	b.WriteString(fmt.Sprintf("           %d autoregs:\n", len(a.AutoRegs)))
	for _, autoreg := range a.AutoRegs {
		b.WriteString(fmt.Sprintf("           Autoregistration: "))
		b.WriteString(fmt.Sprintf("%s ", autoreg))
	}
	fmt.Printf("             Limits: MaxMemory: %dMb\n", a.Limits.MaxMemory)
	return fmt.Sprintf("%s\n", b.String())
}
