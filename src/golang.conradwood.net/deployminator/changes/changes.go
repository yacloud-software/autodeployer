package changes

// this package suggests changes
import (
	//	"context"
	"fmt"
	//	pb "golang.conradwood.net/apis/deployminator"
	"golang.conradwood.net/deployminator/db"
	"golang.conradwood.net/deployminator/targets"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
	"time"
)

var (
	change_trigger_chan = make(chan *change_trigger)
)

type change_trigger struct {
}

func init() {
	go change_scanner_loop()
}
func change_scanner_loop() {
	fmt.Printf("Starting change scanner loop...\n")
	// wait for startup
	sleep := time.Duration(2) * time.Second
	for {
		var ct *change_trigger
		ct = nil
		select {
		case ct = <-change_trigger_chan:
		//
		case <-time.After(sleep):
			//
		}
		if !targets.HasScanned() {
			fmt.Printf("Cannot scan until targets are scanned...\n")
			continue
		}
		sleep = time.Duration(15) * time.Second
		err := scan_for_change(ct)
		if err != nil {
			fmt.Printf("Change scanner failed: %s\n", utils.ErrorString(err))
		}
	}
}

// change_trigger might be nil (if it is a periodically scan)
func scan_for_change(req *change_trigger) error {
	ctx := authremote.Context()
	dds, err := db.Descriptordb.All(ctx)
	if err != nil {
		return err
	}
	var changes []*Change
	for _, dd := range dds {
		if !dd.DeployMe {
			continue
		}
		ctx := authremote.Context()
		fdd, err := db.FetchFull(ctx, dd)
		if err != nil {
			return err
		}
		chs, err := find_need_starting(fdd)
		if err != nil {
			return err
		}
		changes = append(changes, chs...)
	}
	for _, c := range changes {
		fmt.Printf("Change: %s\n", c.String())
	}
	return nil
}
