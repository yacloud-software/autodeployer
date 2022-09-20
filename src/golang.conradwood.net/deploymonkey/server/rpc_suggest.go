package main

import (
	"context"
	"fmt"
	"golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/config"
	"golang.conradwood.net/deploymonkey/suggest"
	"golang.conradwood.net/go-easyops/client"
)

func (d *DeployMonkey) GetSuggestions(ctx context.Context, req *pb.SuggestRequest) (*pb.SuggestionList, error) {
	depl := pb.NewDeployMonkeyClient(client.Connect("deploymonkey.DeployMonkey"))
	depls, err := d.GetDeploymentsFromCache(ctx, &common.Void{})
	if err != nil {
		return nil, fmt.Errorf("Failed to get deployments from cache: %s", err)
	}

	cfg, err := config.GetConfig(depl)
	if err != nil {
		return nil, fmt.Errorf("Could not get config: %s", err)
	}
	s, err := suggest.Analyse(cfg, depls)
	if err != nil {
		return nil, fmt.Errorf("Suggestion failed: %s", err)
	}
	fmt.Println(s.String())
	res := &pb.SuggestionList{}
	for _, ac := range s.Starts {
		sd := &pb.Suggestion{
			Start:         true,
			Host:          ac.Host,
			App:           ac.App,
			DeployRequest: ac.DeployRequest(),
		}
		res.Suggestions = append(res.Suggestions, sd)
	}
	for _, ac := range s.Stops {
		sd := &pb.Suggestion{
			Start:           false,
			Host:            ac.Host,
			App:             ac.App,
			UndeployRequest: ac.UndeployRequest(),
		}
		res.Suggestions = append(res.Suggestions, sd)
	}
	return res, nil
}
