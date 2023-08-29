package common

import (
	"errors"
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var (
	parse_strict = flag.Bool("parse_strict", false, "should be true, but during transition from repository to repositoryid it cannot")
	parse_repo   = flag.Int64("parse_repoid", 0, "if deploy.yaml has no repositoryid, then use this repoid as default. useful for testing only")
)

type FileDef struct {
	Namespace string
	Groups    []*pb.GroupDefinitionRequest
}

func PrintSample() {
	foo := &FileDef{Namespace: "example",
		Groups: []*pb.GroupDefinitionRequest{
			&pb.GroupDefinitionRequest{
				Namespace: "example",
				GroupID:   "foogroupid",
				Applications: []*pb.ApplicationDefinition{
					&pb.ApplicationDefinition{
						Args:      []string{"arg1", "arg2"},
						Limits:    &pb.Limits{MaxMemory: 3000},
						Container: &pb.ContainerDef{URL: "foo"},
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
	res, err := ParseConfig(fb, repoid)
	if err != nil {
		fmt.Printf("Failed to parse file %s: %s\n", fname, err)
		return nil, err
	}
	fmt.Printf("Found %d groups in file %s\n", len(res.Groups), fname)
	fmt.Printf("Namespace: %s\n", res.Namespace)
	for _, x := range res.Groups {
		PrintGroup(x)
	}
	return res, nil
}
func ParseConfig(config []byte, repoid uint64) (*FileDef, error) {
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
			if app.RepositoryID == 0 {
				app.RepositoryID = uint64(*parse_repo)
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
	if app.ArtefactID == 0 {
		if app.RepositoryID == 0 {
			return fmt.Errorf("Missing RepositoryID and ArtefactID")
		}
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
