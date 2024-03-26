package main

import (
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
	sb "golang.conradwood.net/apis/slackgateway"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/client"
	"strings"
)

const (
	NOTIFY_ON_DEPLOY = false
)

var (
	slack sb.SlackGatewayClient
)

func NotifyPeopleAboutDeploy(apps []*pb.ApplicationDefinition, version int) {
	if !NOTIFY_ON_DEPLOY {
		return
	}
	var err error
	if slack == nil {
		slack = sb.NewSlackGatewayClient(client.Connect("slackgateway.SlackGateway"))
	}
	ctx := authremote.Context()
	msg := fmt.Sprintf("Datacenter update:\nApplied change #%d, containing: \n", version)
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

func NotifyPeopleAboutCancel(sr *stopRequest, usermessages []string, emsg string) {
	var err error
	if slack == nil {
		slack = sb.NewSlackGatewayClient(client.Connect("slackgateway.SlackGateway"))
	}
	name := "unknown"
	if sr.deployInfo != nil {
		x := fmt.Sprintf("(%s)", sr.deployInfo.Namespace)
		if sr.deployInfo.AppReference != nil && sr.deployInfo.AppReference.AppDef != nil {
			ad := sr.deployInfo.AppReference.AppDef
			x = fmt.Sprintf("(%s)", ad.Binary)
		}
		name = fmt.Sprintf("Repository #%d%s", sr.deployInfo.RepositoryID, x)
	}
	ctx := authremote.Context()
	msg := fmt.Sprintf("Datacenter update of \"%s\" cancelled: %s\n", name, emsg)
	for _, umsg := range usermessages {
		umsg = strings.TrimSuffix(umsg, "\n")
		msg = msg + umsg + "\n"
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
