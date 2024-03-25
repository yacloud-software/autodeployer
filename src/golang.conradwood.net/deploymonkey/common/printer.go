package common

import (
	"bytes"
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
)

type PDBGroup interface {
}

func PrintGroup(dbg PDBGroup) {
	fmt.Printf("DBGROUP: %#v\n", dbg)
}

/*
	func PrintGroup(x *pb.GroupDefinitionRequest) {
		fmt.Printf("  Group: %s with %d applications\n", x.GroupID, len(x.Applications))
		fmt.Printf("        Namespace  : %s\n", x.Namespace)
		for _, a := range x.Applications {
			fmt.Printf("%s", PrintApp(a))
		}
	}
*/
func PrintApp(a *pb.ApplicationDefinition) string {
	var b bytes.Buffer

	fmt.Printf("        Application: (AlwaysOn: %v, Critical %v)\n", a.AlwaysOn, a.Critical)
	fmt.Printf("             Repo  : %d\n", a.RepositoryID)
	fmt.Printf("             Binary: %s\n", a.Binary)
	fmt.Printf("            BuildID: %d\n", a.BuildID)
	fmt.Printf("           Machines: %s\n", a.Machines)
	fmt.Printf("               Type: %s\n", a.DeployType)
	fmt.Printf("             Public: %v\n", a.Public)
	b.WriteString(fmt.Sprintf("           %d Args: ", len(a.Args)))
	for _, arg := range a.Args {
		b.WriteString(fmt.Sprintf("%s ", arg))
	}
	b.WriteString("%\n")
	b.WriteString(fmt.Sprintf("           %d autoregs:\n", len(a.AutoRegs)))
	for _, autoreg := range a.AutoRegs {
		b.WriteString(fmt.Sprintf("           Autoregistration: "))
		b.WriteString(fmt.Sprintf("%s ", autoreg))
	}
	fmt.Printf("             Limits: MaxMemory: %dMb\n", a.Limits.MaxMemory)
	return fmt.Sprintf("%s\n", b.String())
}
