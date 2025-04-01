package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	common "golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/errors"
)

const (
	apply_timeout = time.Duration(120) * time.Second
)

var (
	apply_suggest_lock sync.Mutex
)

func (dm *DeployMonkey) ApplySuggestions(ctx context.Context, req *common.Void) (*pb.SuggestionList, error) {
	sl, err := dm.GetSuggestions(ctx, &pb.SuggestRequest{})
	if err != nil {
		return nil, err
	}
	dm.triggerApplySuggestions(sl)
	return sl, nil
}
func (dm *DeployMonkey) triggerApplyAllSuggestions() {
	ctx := authremote.ContextWithTimeout(apply_timeout)
	sl, err := dm.GetSuggestions(ctx, &pb.SuggestRequest{})
	if err != nil {
		fmt.Printf("Failed to get suggestions: %s\n", errors.ErrorString(err))
		return
	}
	dm.triggerApplySuggestions(sl)
}
func (dm *DeployMonkey) triggerApplySuggestions(sl *pb.SuggestionList) {
	go dm.applySuggestions(sl)
}

func (dm *DeployMonkey) applySuggestions(sl *pb.SuggestionList) error {
	apply_suggest_lock.Lock()
	defer apply_suggest_lock.Unlock()
	//	ctx := authremote.Context()
	if len(sl.Suggestions) == 0 {
		fmt.Printf("No suggestions to apply\n")
		return nil
	}
	err := dm.try_suggestions(sl)
	if err != nil {
		fmt.Printf("Failed to deploy: %s\n", err)
		return err
	}

	return nil
}

func (dm *DeployMonkey) try_suggestions(s *pb.SuggestionList) error {
	var err error
	fmt.Printf("Executing %d requests...\n", len(s.Suggestions))
	for _, start := range s.Suggestions {
		if !start.Start {
			continue
		}
		//		fmt.Println(Suggestion2Line(start))
		ctx := authremote.ContextWithTimeout(apply_timeout)
		fmt.Printf("Deploying %s...\n", start.String())
		d := dm.ToDeployRequest(start)
		_, err = dm.DeployAppOnTarget(ctx, d)
		if err != nil {
			return err
		}

	}
	fmt.Printf("Executing stop requests...\n")
	for _, stop := range s.Suggestions {
		if stop.Start {
			continue
		}
		//		fmt.Println(Suggestion2Line(stop))
		d := dm.ToUndeployRequest(stop)
		ctx := authremote.ContextWithTimeout(apply_timeout)
		fmt.Printf("Undeploying %s...\n", stop.String())
		_, err = dm.UndeployAppOnTarget(ctx, d)
		if err != nil {
			return err
		}
	}
	return nil
}
func (dm *DeployMonkey) ToDeployRequest(suggestion *pb.Suggestion) *pb.DeployAppRequest {
	return suggestion.DeployRequest
}
func (dm *DeployMonkey) ToUndeployRequest(suggestion *pb.Suggestion) *pb.UndeployAppRequest {
	return suggestion.UndeployRequest
}
