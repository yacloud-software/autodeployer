package main

import (
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	sb "golang.conradwood.net/apis/slackgateway"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/tokens"
)

const (
	NOTIFY_ON_DEPLOY = false
)

var (
	slack sb.SlackGatewayClient
)

func NotifyPeopleAboutDeploy(dbgroup *DBGroup, apps []*pb.ApplicationDefinition, version int) {
	if !NOTIFY_ON_DEPLOY {
		return
	}
	var err error
	if slack == nil {
		slack = sb.NewSlackGatewayClient(client.Connect("slackgateway.SlackGateway"))
	}
	ctx := tokens.ContextWithToken()
	msg := fmt.Sprintf("Datacenter update:\nApplied change #%d (%s), containing: \n", version, dbgroup.groupDef.Namespace)
	for _, app := range apps {
		msg = msg + fmt.Sprintf("   %d instances: build #%d of application %s\n", app.Instances, app.BuildID, app.Binary)
	}
	fmt.Printf("slack: Posting message: %s\n", msg)
	pm := &sb.PublishMessageRequest{OriginService: "originservicenotfilledinyet",
		Channel: "deployments",
		Text:    msg,
	}
	_, err = slack.PublishMessage(ctx, pm)
	if err != nil {
		fmt.Printf("Failed to post slack message: %s\n", err)
	}

}

func NotifyPeopleAboutCancel(sr *stopRequest, emsg string) {
	var err error
	if slack == nil {
		slack = sb.NewSlackGatewayClient(client.Connect("slackgateway.SlackGateway"))
	}
	name := "unknown"
	if sr.deployInfo != nil {
		name = fmt.Sprintf("Repository #%d", sr.deployInfo.RepositoryID)
	}
	ctx := tokens.ContextWithToken()
	msg := fmt.Sprintf("Datacenter update of \"%s\" cancelled: %s\n", name, emsg)
	fmt.Printf("slack: Posting message: %s\n", msg)
	pm := &sb.PublishMessageRequest{OriginService: "originservicenotfilledinyet",
		Channel: "deployments",
		Text:    msg,
	}
	_, err = slack.PublishMessage(ctx, pm)
	if err != nil {
		fmt.Printf("Failed to post slack message: %s\n", err)
	}

}
