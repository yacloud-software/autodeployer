package deployq

import (
	pb "golang.conradwood.net/apis/deploymonkey"
	dp "golang.conradwood.net/deploymonkey/deployplacements"
	"strings"
)

var (
	// if deployed with "per instance", +1 will be added to the score
	bin_score_match = map[string]int{
		"secureargs-server":  20,
		"logservice-server":  10,
		"errorlogger-server": 10,
		"objectauth-server":  10,
		"objectstore-server": 10,
	}
)

type deployTransaction struct {
	requests []*dp.DeployRequest
}

func (dt *deployTransaction) Score() int {
	has_instances := false
	app_score := 0
	for _, r := range dt.requests {
		appdef := r.AppDef()
		if appdef.InstancesMeansPerAutodeployer {
			has_instances = true
		}
		as := appScore(appdef)
		if as > app_score {
			app_score = as
		}
	}
	res := app_score
	if has_instances {
		res++
	}
	return res
}

func appScore(ad *pb.ApplicationDefinition) int {
	for k, v := range bin_score_match {
		if strings.Contains(ad.Binary, k) {
			return v
		}
	}
	return 0
}
