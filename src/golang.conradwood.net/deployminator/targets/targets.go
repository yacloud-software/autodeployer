package targets

import (
	"flag"
	"fmt"
	reg "golang.conradwood.net/apis/registry"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/utils"
	"sync"
	"time"
)

var (
	debug         = flag.Bool("debug_targets", false, "debug target discovery")
	scan_complete bool
	targets       []*Target
	tlock         sync.Mutex
)

func init() {
	scan_complete = false
	go target_scan_loop()
}
func target_scan_loop() {
	time.Sleep(5 * time.Second)
	rc := client.GetRegistryClient()
	for {
		treq := &reg.V2GetTargetRequest{ApiType: reg.Apitype_grpc, ServiceName: []string{"autodeployer.AutoDeployer"}}
		ctx := authremote.Context()
		tr, err := rc.V2GetTarget(ctx, treq)
		if err != nil {
			fmt.Printf("Error scanning registry: %s\n", utils.ErrorString(err))
			time.Sleep(5 * time.Second)
			continue
		}
		wg := sync.WaitGroup{}
		for _, t := range tr.Targets {
			adr := fmt.Sprintf("%s:%d", t.IP, t.Port)
			ac := false
			for _, apitype := range t.ApiType {
				if apitype == reg.Apitype_grpc {
					ac = true
					break
				}
			}
			if !ac {
				continue
			}
			wg.Add(1)
			go func(a string) {
				err = scan(a)
				if err != nil {
					fmt.Printf("Failed to scan target %s: %s\n", a, utils.ErrorString(err))
				}
				wg.Done()
			}(adr)
		}
		wg.Wait()
		if !scan_complete {
			fmt.Printf("First scan complete.\n")
		}
		scan_complete = true
		time.Sleep(30 * time.Second)

	}
}
func HasScanned() bool {
	return scan_complete
}
func scan(address string) error {
	if *debug {
		fmt.Printf("Scanning %s\n", address)
	}
	t := findTargetByAddress(address)
	return t.Scan()
}
func findTargetByAddress(address string) *Target {
	tlock.Lock()
	defer tlock.Unlock()
	for _, t := range targets {
		if t.address == address {
			return t
		}
	}
	t := &Target{address: address}
	targets = append(targets, t)
	return t
}

func GetMatchList() *MatchList {
	ml := &MatchList{}
	for _, t := range targets {
		for _, app := range t.apps {
			ml.add(t, app)
		}
	}
	return ml
}

func GetTargets() *TargetList {
	return &TargetList{targets: targets}
}
