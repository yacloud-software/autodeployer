package main

import (
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/autodeployer"
	//	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/utils"
	"strings"
)

func main() {
	flag.Parse()
	for _, ad := range ad_list() {
		Update(ad)
	}
}

func ad_list() []string {
	if len(flag.Args()) == 0 {
		// read reg
		panic("cannot read registry yet, please use args")
	}
	return flag.Args()
}

func Update(host string) error {
	fmt.Printf("Updating %s\n", host)
	DownloadLatest()
	err := TransferAD(host)
	if err != nil {
		fmt.Printf("Failed to update %s: %s\n", host, err)
		return err
	}

	if !strings.Contains(host, ":") {
		host = fmt.Sprintf("%s:4000", host)
	}
	fmt.Printf("Connecting to server %s\n", host)
	conn, err := client.ConnectWithIP(host)
	if err != nil {
		return err
	}
	defer conn.Close()
	cl := pb.NewAutoDeployerClient(conn)
	ctx := authremote.Context()
	_, err = cl.StopAutodeployer(ctx, &pb.StopRequest{})
	if err != nil {
		return err
	}
	return nil
}

func DownloadLatest() {
	if utils.FileExists("/tmp/autodeployer-server") {
		return
	}
	panic("cannot download latest autodeployer-server yet")
}

// replace "/usr/local/bin/autodeployer-server" with latest binary on "host"
func TransferAD(host string) error {
	l := linux.New()
	out, err := l.SafelyExecute([]string{"scp", "/tmp/autodeployer-server", fmt.Sprintf("%s:/tmp/", host)}, nil)
	if err != nil {
		fmt.Printf("Failed to copy: %s\n", out)
		return err
	}

	l = linux.New()
	out, err = l.SafelyExecute([]string{"ssh", host, "bash", "-c", "sudo rm /usr/local/bin/autodeployer-server ; sudo chmod 755 /tmp/autodeployer-server ; sudo mv /tmp/autodeployer-server /usr/local/bin/"}, nil)
	if err != nil {
		fmt.Printf("Failed to overwrite: %s\n", out)
		return err
	}
	return nil
}
