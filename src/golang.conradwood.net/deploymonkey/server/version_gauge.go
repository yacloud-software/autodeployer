package main

import (
    "golang.conradwood.net/go-easyops/authremote"
	"flag"
	"fmt"
	"golang.conradwood.net/go-easyops/prometheus"
	"time"
	//	pb "golang.conradwood.net/apis/deploymonkey"
)

var (
	softwareGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "deploymonkey_software_versions",
			Help: "versions of software",
		},
		[]string{"namespace", "repositoryid", "group"},
	)
	version_gauge_interval = flag.Int("version_gauge_interval", 120, "in `seconds`: how often to update the version gauge")
)

func startVersionGauge() {
	prometheus.MustRegister(softwareGauge)
	fmt.Printf("Starting version gauge...\n")
	updateVersionGauge()
	ticker := time.NewTicker(time.Duration(*version_gauge_interval) * time.Second)
	go func() {
		for _ = range ticker.C {
			updateVersionGauge()
		}
	}()
}

func updateVersionGauge() {
	if *debug {
		fmt.Printf("updating version gauge...\n")
	}
	s := new(DeployMonkey)
	cfg, err := s.GetConfig(authremote.Context(), nil)
	if err != nil {
		fmt.Printf("Failed to update version gauge: %s\n", err)
		return
	}
	for _, config := range cfg.GroupConfigs {
		for _, app := range config.Applications {
			l := prometheus.Labels{
				"namespace":    config.Group.NameSpace,
				"group":        config.Group.GroupID,
				"repositoryid": fmt.Sprintf("%d", app.RepositoryID),
			}
			softwareGauge.With(l).Set(float64(app.BuildID))
		}
	}

}
