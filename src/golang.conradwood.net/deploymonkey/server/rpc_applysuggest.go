package main

import (
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
	apply_suggest_chan   = make(chan *apply_suggest_event)
	apply_suggest_lock   sync.Mutex
	applying_suggestions = false
)

type apply_suggest_event struct {
	suggestion_list *pb.SuggestionList
}

func init() {
	go apply_suggest_thread()
}

func (dm *DeployMonkey) ApplySuggestions(req *common.Void, srv pb.DeployMonkey_ApplySuggestionsServer) error {
	if applying_suggestions {
		return errors.Errorf("Already applying. retry later")
	}
	ctx := srv.Context()
	sl, err := dm.GetSuggestions(ctx, &pb.SuggestRequest{})
	if err != nil {
		return err
	}
	dm.triggerApplySuggestions(sl)
	return nil
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
	apply_suggest_chan <- &apply_suggest_event{suggestion_list: sl}
	return nil
}

func try_suggestions(s *pb.SuggestionList) error {
	applying_suggestions = true
	defer func() {
		applying_suggestions = false
	}()

	var err error
	fmt.Printf("Executing %d requests...\n", len(s.Suggestions))
	dm := &DeployMonkey{}
	for _, start := range s.Suggestions {
		if !start.Start {
			continue
		}
		//		fmt.Println(Suggestion2Line(start))
		ctx := authremote.ContextWithTimeout(apply_timeout)
		fmt.Printf("Deploying %s...\n", start.String())
		d := ToDeployRequest(start)
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
		d := ToUndeployRequest(stop)
		ctx := authremote.ContextWithTimeout(apply_timeout)
		fmt.Printf("Undeploying %s...\n", stop.String())
		_, err = dm.UndeployAppOnTarget(ctx, d)
		if err != nil {
			return err
		}
	}
	return nil
}
func ToDeployRequest(suggestion *pb.Suggestion) *pb.DeployAppRequest {
	return suggestion.DeployRequest
}
func ToUndeployRequest(suggestion *pb.Suggestion) *pb.UndeployAppRequest {
	return suggestion.UndeployRequest
}

func apply_suggest_thread() {
	for {
		ase := <-apply_suggest_chan
		try_suggestions(ase.suggestion_list)
	}
}
