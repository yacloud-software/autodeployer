package config

/*
this file contains the "nftables" code.
the iptables code and the flag/switches to determine wether to use iptables or nftables
are in the fil action_port.go

this code maintains a complete list of mappings and applies a complete set of nftables rules in nat autodeployer chain
whenever they change
*/

import (
	"bytes"
	"flag"
	"fmt"
	"golang.conradwood.net/autodeployer/deployments"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/utils"
	"sync"
	tt "text/template"
)

const (
	nf_templ1 = `
table {{.Family}} nat {
        chain autodeployer_nat {
{{range .Ports}} tcp dport {{.PublicPort}} dnat to :{{.ModulePort}}
{{end}}
        }
}
`
	nf_templ2 = `
table {{.Family}} filter {
        chain autodeployer_filter {
{{range .Ports}} tcp dport {{.ModulePort}} mark set 412 accept
{{end}}
        }
}

`
)

var (
	nft_family = flag.String("nft_family", "inet", "usually either ip or inet")
	//	dir        = flag.String("template_dir", "/etc/cnw/autodeployer/templates/", "directory with templates")
	nf_portmap = make(map[int]int) // module's port(key) -> public port(value)
	nf_lock    sync.Mutex
)

type TemplateData struct {
	Family string
	Ports  []*TemplatePortData
}
type TemplatePortData struct {
	ModulePort int
	PublicPort int
}

type NFPortAction struct {
	du  *deployments.Deployed
	cfg *ApplicationPort
}

func NewNFPortAction(ap *ApplicationPort, du *deployments.Deployed) (Action, error) {
	return &NFPortAction{du: du, cfg: ap}, nil

}

func (pa *NFPortAction) ID() string {
	return pa.du.StartupMsg
}

func (pa *NFPortAction) Apply() error {
	nf_lock.Lock()
	defer nf_lock.Unlock()

	if pa.cfg.PortIndex < 1 {
		return fmt.Errorf("Port cannot be less than 1 (%d)", pa.cfg.PortIndex)
	}
	if len(pa.du.Ports) < pa.cfg.PortIndex {
		fmt.Printf("nftables Error, listing ports now:\n")
		for i, p := range pa.du.Ports {
			fmt.Printf("%d. Port: %d\n", i, p)
		}
		return fmt.Errorf("No such port #%d (got only %d ports)", pa.cfg.PortIndex, len(pa.du.Ports))
	}
	/* activate the following mapping:
	Connections to public port (pa.cfg.PublicPort)
	are redirected to
	module's port (pa.du.Ports[pa.cfg.PortIndex[pa.cfg.PortIndex-1])
	*/
	nf_portmap[pa.du.Ports[pa.cfg.PortIndex-1]] = pa.cfg.PublicPort
	err := pa.exe()
	if err != nil {
		fmt.Printf("Failed to apply nftables port action %s: %s\n", pa.String(), err)
		return err
	}
	fmt.Printf("Applied nftables port action: %s\n", pa.String())
	return nil
}
func (pa *NFPortAction) Unapply() error {
	nf_lock.Lock()
	defer nf_lock.Unlock()

	if pa.cfg.PortIndex < 1 {
		return fmt.Errorf("Port cannot be less than 1 (%d)", pa.cfg.PortIndex)
	}
	if len(pa.du.Ports) < pa.cfg.PortIndex {
		return fmt.Errorf("No such port #%d", pa.cfg.PortIndex)
	}
	delete(nf_portmap, pa.du.Ports[pa.cfg.PortIndex-1])
	err := pa.exe()
	if err != nil {
		fmt.Printf("Failed to unapply nftables port action %s: %s\n", pa.String(), err)
		return err
	}
	fmt.Printf("Unapplied nftables port action: %s\n", pa.String())
	return err
}
func (pa *NFPortAction) String() string {
	return fmt.Sprintf("%d->%d", pa.cfg.PortIndex, pa.cfg.PublicPort)
}

func ResetNFPorts() {
	nf_lock.Lock()
	defer nf_lock.Unlock()
	if *debug_apply {
		fmt.Printf("Resetting ports...\n")
	}
	out, err := linux.New().SafelyExecute([]string{"/usr/sbin/nft", "flush", "chain", *nft_family, "nat", "autodeployer_nat"}, nil)
	if err != nil {
		fmt.Printf("(2) nftables flush failed: %s\n%s\n", out, err)
	}
	out, err = linux.New().SafelyExecute([]string{"/usr/sbin/nft", "flush", "chain", *nft_family, "filter", "autodeployer_filter"}, nil)
	if err != nil {
		fmt.Printf("(3) nftables flush failed: %s\n%s\n", out, err)
	}
}

func getNFTemplates() []string {
	return []string{nf_templ1, nf_templ2}
}

func (pa *NFPortAction) exe() error {
	out, err := linux.New().SafelyExecute([]string{"/usr/sbin/nft", "flush", "chain", *nft_family, "nat", "autodeployer_nat"}, nil)
	if err != nil {
		fmt.Printf("(1) nftables flush failed: %s\n%s\n", out, err)
		// must continue here. because it might just be a conditrion where table does not exist
		//		return err
	}
	out, err = linux.New().SafelyExecute([]string{"/usr/sbin/nft", "flush", "chain", *nft_family, "filter", "autodeployer_filter"}, nil)
	if err != nil {
		fmt.Printf("(1) nftables flush failed: %s\n%s\n", out, err)
		// must continue here. because it might just be a conditrion where table does not exist
		//		return err
	}
	for i, t := range getNFTemplates() {
		template := tt.New(fmt.Sprintf("nftables-%d", i))
		_, err := template.Parse(t)
		if err != nil {
			fmt.Printf("nftables action_port2.go: Template broken! (%s)\n", err)
			return err
		}
		data := &TemplateData{Family: *nft_family}
		for k, v := range nf_portmap {
			tp := &TemplatePortData{ModulePort: k, PublicPort: v}
			data.Ports = append(data.Ports, tp)
		}
		b := &bytes.Buffer{}
		err = template.Execute(b, data)
		if err != nil {
			return err
		}
		err = utils.WriteFile("/tmp/autodeployer_nftables.tmp", b.Bytes())
		if err != nil {
			return err
		}

		out, err = linux.New().SafelyExecute([]string{"/usr/sbin/nft", "-f", "/tmp/autodeployer_nftables.tmp"}, nil)
		if err != nil {
			fmt.Printf("nftables Failed 2: %s\n%s\n", out, err)
			return err
		}
	}
	return nil

}
