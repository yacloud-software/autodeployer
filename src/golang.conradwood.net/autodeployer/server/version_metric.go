package main

import (
	"fmt"
	"golang.conradwood.net/go-easyops/prometheus"
	"golang.conradwood.net/go-easyops/utils"
	"strconv"
	"strings"
	"time"
)

var (
	versionGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "autodeployer_debianversion",
			Help: "V=1 UNIT=none DESC=value in /etc/debian_version",
		})
)

func version_metric() {
	prometheus.MustRegister(versionGauge)
	t := time.Duration(5) * time.Second
	for {
		time.Sleep(t)
		t = time.Duration(120) * time.Second
		bs, err := utils.ReadFile("/etc/debian_version")
		if err != nil {
			fmt.Printf("cannot read file: %s\n", err)
			continue
		}
		s := string(bs)
		s = strings.Trim(s, "\n")
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			fmt.Printf("cannot parse number \"%s\": %s\n", s, err)
			continue
		}
		//	fmt.Printf("Debian Version: %0.2f\n", f)
		versionGauge.Set(f)
	}
}
