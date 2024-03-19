package main

import (
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
)

func makeitso_new(group *DBGroup, apps []*pb.ApplicationDefinition) error {
	fmt.Printf("meant to deploy %d apps, but cannot, because it is too new\n", len(apps))
	return nil
}
