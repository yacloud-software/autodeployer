package main

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/deployminator"
	"golang.conradwood.net/deployminator/config"
)

// the old-style config mangled all these together. we're parsing it and passing it out
type CompatApp struct {
	app          *pb.Application
	args         []string
	machinegroup string
	critical     bool
	instances    uint32
}

func findOrCreateAppsFromOldConfig(ctx context.Context, repoid uint64, config *config.FileDef) ([]*CompatApp, error) {
	var res []*CompatApp
	for _, g := range config.Groups {
		for _, a := range g.Applications {
			fmt.Printf("Binary: %s, URL: %s\n", a.Binary, a.DownloadURL)
			app, err := findOrCreateApp(ctx, repoid, a.Binary, a.DownloadURL)
			if err != nil {
				return nil, err
			}
			ca := &CompatApp{
				app:          app,
				args:         a.Args,
				instances:    a.Instances,
				machinegroup: a.Machines,
				critical:     a.Critical,
			}
			res = append(res, ca)
		}
	}
	return res, nil
}
