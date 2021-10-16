package config

/*
this file contains the "iptables" code and the flag/switches to determine wether to use iptables or nftables.
the nftables code itself is in file action_port2.go

this code adds/removes rules from the "-t nat autodeployer" chain as necessary
*/

import (
	"flag"
	"fmt"
	"golang.conradwood.net/autodeployer/deployments"
	"golang.conradwood.net/autodeployer/linux"
)

var (
	use_nftables = flag.Bool("use_nftables", true, "if false, it uses the old iptables code instead")
)

func NewPortAction(ap *ApplicationPort, du *deployments.Deployed) (Action, error) {
	if *use_nftables {
		return NewNFPortAction(ap, du)
	}
	/*
		if len(du.Ports) == 0 {
			panic(fmt.Sprintf("No ports in %s!", du.String()))
		}
	*/
	return &PortAction{du: du, cfg: ap}, nil

}

type PortAction struct {
	du  *deployments.Deployed
	cfg *ApplicationPort
}

func (pa *PortAction) ID() string {
	return pa.du.StartupMsg
}

func (pa *PortAction) Apply() error {
	if pa.cfg.PortIndex < 1 {
		return fmt.Errorf("Port cannot be less than 1 (%d)", pa.cfg.PortIndex)
	}
	sp := fmt.Sprintf("%d", pa.cfg.PublicPort)
	if len(pa.du.Ports) < pa.cfg.PortIndex {
		fmt.Printf("Error, listing ports now:\n")
		for i, p := range pa.du.Ports {
			fmt.Printf("%d. Port: %d\n", i, p)
		}
		return fmt.Errorf("No such port #%d (got only %d ports)", pa.cfg.PortIndex, len(pa.du.Ports))
	}
	dp := fmt.Sprintf(":%d", pa.du.Ports[pa.cfg.PortIndex-1])
	out, err := linux.SafelyExecute([]string{"/sbin/iptables", "-t", "nat", "-A", "autodeployer", "-i", "eth0", "-p", "tcp", "--destination-port", sp, "-j", "DNAT", "--to-destination", dp})
	if err != nil {
		fmt.Printf("failed to apply %s: %s (%s)\n", pa.String(), out, err)
	}
	out, err = linux.SafelyExecute([]string{"/sbin/iptables", "-A", "autodeployer", "-i", "eth0", "-p", "tcp", "--destination-port", dp, "-j", "ACCEPT"})
	if err != nil {
		fmt.Printf("failed to apply %s: %s (%s)\n", pa.String(), out, err)
	}
	fmt.Printf("Applied iptables port action: %s\n", pa.String())
	return nil
}
func (pa *PortAction) Unapply() error {
	if pa.cfg.PortIndex < 1 {
		return fmt.Errorf("Port cannot be less than 1 (%d)", pa.cfg.PortIndex)
	}
	sp := fmt.Sprintf("%d", pa.cfg.PublicPort)
	if len(pa.du.Ports) < pa.cfg.PortIndex {
		return fmt.Errorf("No such port #%d", pa.cfg.PortIndex)
	}
	dp := fmt.Sprintf(":%d", pa.du.Ports[pa.cfg.PortIndex-1])

	out, err := linux.SafelyExecute([]string{"/sbin/iptables", "-D", "autodeployer", "-i", "eth0", "-p", "tcp", "--destination-port", dp, "-j", "ACCEPT"})
	if err != nil {
		fmt.Printf("failed to apply %s: %s (%s)\n", pa.String(), out, err)
	}

	out, err = linux.SafelyExecute([]string{"/sbin/iptables", "-t", "nat", "-D", "autodeployer", "-i", "eth0", "-p", "tcp", "--destination-port", sp, "-j", "DNAT", "--to-destination", dp})
	if err != nil {
		fmt.Printf("failed to unapply %s: %s (%s)\n", pa.String(), out, err)
	}
	fmt.Printf("Unapplied iptables port action: %s\n", pa.String())
	return nil
}
func (pa *PortAction) String() string {
	return fmt.Sprintf("%d->%d", pa.cfg.PortIndex, pa.cfg.PublicPort)
}

func ResetPorts() {
	if *use_nftables {
		ResetNFPorts()
		return
	}
	linux.SafelyExecute([]string{"/sbin/iptables", "-t", "nat", "-N", "autodeployer"})
	linux.SafelyExecute([]string{"/sbin/iptables", "-t", "nat", "-F", "autodeployer"})
	linux.SafelyExecute([]string{"/sbin/iptables", "-N", "autodeployer"})
	linux.SafelyExecute([]string{"/sbin/iptables", "-F", "autodeployer"})
}
