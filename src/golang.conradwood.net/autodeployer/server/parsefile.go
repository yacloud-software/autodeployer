package main

import (
	"errors"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type FileDef struct {
	Namespace string
	Groups    []*pb.GroupDefinitionRequest
}

func ParseFile(fname string) (*FileDef, error) {
	fmt.Printf("Parsing %s\n", fname)
	fb, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Printf("Failed to read file %s: %s\n", fname, err)
		return nil, err
	}
	gd := FileDef{}
	err = yaml.Unmarshal(fb, &gd)
	if err != nil {
		fmt.Printf("Failed to parse file %s: %s\n", fname, err)
		return nil, err
	}
	// apply namespace throughout
	for _, x := range gd.Groups {
		if x.Namespace == "" {
			x.Namespace = gd.Namespace
		}
	}
	for _, g := range gd.Groups {
		for _, app := range g.Applications {
			err = CheckAppComplete(app)
			if err != nil {
				return nil, err
			}
		}
	}
	fmt.Printf("Found %d groups in file %s\n", len(gd.Groups), fname)
	fmt.Printf("Namespace: %s\n", gd.Namespace)
	return &gd, nil
}
func CheckAppComplete(app *pb.ApplicationDefinition) error {
	s := fmt.Sprintf("%d-%s", app.RepositoryID, app.DeploymentID)
	if app.DownloadURL == "" {
		return errors.New(fmt.Sprintf("%s is invalid: Missing DownloadURL", s))
	}
	return nil
}
