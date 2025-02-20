package main

import (
	"context"
	"fmt"
	"time"

	"golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/config"
	"golang.conradwood.net/deploymonkey/suggest"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/errors"
)

var (
	last_suggestion_list_time time.Time
	last_suggestion_list      *pb.SuggestionList
)

func (d *DeployMonkey) GetSuggestions(ctx context.Context, req *pb.SuggestRequest) (*pb.SuggestionList, error) {
	if last_suggestion_list != nil && time.Since(last_suggestion_list_time) < time.Duration(5)*time.Second {
		return last_suggestion_list, nil
	}
	res, err := d.getSuggestions(ctx, req)
	if err != nil {
		return nil, err
	}
	last_suggestion_list = res
	last_suggestion_list_time = time.Now()
	return last_suggestion_list, nil
}
func (d *DeployMonkey) getSuggestions(ctx context.Context, req *pb.SuggestRequest) (*pb.SuggestionList, error) {
	depl := pb.NewDeployMonkeyClient(client.Connect("deploymonkey.DeployMonkey"))
	depls, err := d.GetDeploymentsFromCache(ctx, &common.Void{})
	if err != nil {
		return nil, errors.Errorf("Failed to get deployments from cache: %s", err)
	}

	cfg, err := config.GetConfig(depl)
	if err != nil {
		return nil, errors.Errorf("Could not get config: %s", err)
	}
	s, err := suggest.Analyse(cfg, depls)
	if err != nil {
		return nil, errors.Errorf("Suggestion failed: %s", err)
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
