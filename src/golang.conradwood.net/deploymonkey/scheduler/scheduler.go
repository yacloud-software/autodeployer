package scheduler

// sometimes it's important to track changes in suggestions.
// for example, in order to not get in the way during normal
// deployments we want to be quite hands-off when deployments
// are triggered.
// when a server is rebooted

// this package provides a few very global and defined
// state changes and when they occured.

// we can then trigger actions based on events
// for example:
// we don't blindly re-apply suggestions if
// a human requested deployment shortly before

import (
	"flag"
	"fmt"
	"golang.conradwood.net/go-easyops/prometheus"
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/suggest"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/tokens"
	"golang.conradwood.net/go-easyops/utils"
	"time"
)

var (
	schedStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "deploymonkey_scheduler_status",
			Help: "status of scheduler",
		},
		[]string{},
	)
	depl               pb.DeployMonkeyClient
	DefaultScheduler   Scheduler
	dryrun             = flag.Bool("scheduler_dryrun", false, "scheduler in dry-run mode (will not apply anything)")
	debug_sch          = flag.Bool("debug_scheduler", false, "debug scheduling code")
	minDeployWait      = flag.Int("scheduler_sleeps_deploy", 10, "`minutes` of graceperiod after any deployment for the scheduler")
	minConfigWait      = flag.Int("scheduler_sleeps_config", 10, "`minutes` of graceperiod after any config-change for the scheduler")
	minSuggestionsWait = flag.Int("scheduler_sleeps_suggest", 10, "`minutes` of graceperiod after any suggestion changes for the scheduler")
)

func init() {
	prometheus.MustRegister(schedStatus)
	go DefaultScheduler.loop()
}

type Scheduler struct {
	lastDeployRequested    time.Time
	lastConfigChanged      time.Time
	lastSuggestionsChanged time.Time
	lastSuggestion         *suggest.Suggestion
}

func (s *Scheduler) DeployRequested() {
	s.lastDeployRequested = time.Now()
}
func (s *Scheduler) ConfigChanged() {
	s.lastConfigChanged = time.Now()
}
func (s *Scheduler) SuggestionsChanged(sug *suggest.Suggestion) {
	s.lastSuggestionsChanged = time.Now()
	s.lastSuggestion = sug
}
func (s *Scheduler) loop() {
	for {
		utils.RandomStall(1)
		s.check()
	}
}
func (s *Scheduler) check() {
	if *debug_sch {
		fmt.Printf("[scheduler] checking status...\n")
	}
	if s.lastSuggestion == nil || s.lastSuggestion.Count() == 0 {
		if *debug_sch {
			fmt.Printf("[scheduler] No suggestions currently\n")
		}
		setStatus(0)
		return
	}
	if time.Since(s.lastDeployRequested) <= (time.Duration(*minDeployWait) * time.Minute) {
		if *debug_sch {
			fmt.Printf("[scheduler] blocked because deploy was at %v\n", s.lastDeployRequested)
		}
		setStatus(1)
		return
	}

	if time.Since(s.lastConfigChanged) <= (time.Duration(*minConfigWait) * time.Minute) {
		if *debug_sch {
			fmt.Printf("[scheduler] blocked because config change was at %v\n", s.lastConfigChanged)
		}
		setStatus(2)
		return
	}

	if time.Since(s.lastSuggestionsChanged) <= (time.Duration(*minSuggestionsWait) * time.Minute) {
		if *debug_sch {
			fmt.Printf("[scheduler] blocked because suggestions changed at %v\n", s.lastSuggestionsChanged)
		}
		setStatus(3)
		return
	}
	s.applySuggestions()
}

func (s *Scheduler) applySuggestions() {
	if depl == nil {
		depl = pb.NewDeployMonkeyClient(client.Connect("deploymonkey.DeployMonkey"))
	}
	fmt.Printf("[scheduler] Applying %s\n", s.lastSuggestion.String())
	setStatus(20)
	if *dryrun {
		fmt.Printf("[scheduler] Abort applying - sched_dryrun flag is set\n")
		return
	}
	ctx := tokens.ContextWithToken()
	var haderr error
	for _, start := range s.lastSuggestion.Starts {
		d := start.DeployRequest()
		_, err := depl.DeployAppOnTarget(ctx, d)
		if err != nil {
			haderr = err
			fmt.Printf("Error deploying app: %s\n", err)
		}
	}
	if haderr != nil {
		fmt.Printf("Not undeploying stuff, because we had errors deploying stuff\n")
		return
	}
	for _, start := range s.lastSuggestion.Stops {
		d := start.UndeployRequest()
		_, err := depl.UndeployAppOnTarget(ctx, d)
		if err != nil {
			fmt.Printf("Error undeploying app: %s\n", err)
		}
	}
}

func setStatus(status int) {
	schedStatus.With(prometheus.Labels{}).Set(float64(status))
}
