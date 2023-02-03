package main

import (
	"context"
	"fmt"
	"golang.conradwood.net/apis/common"
	//rpb "golang.conradwood.net/apis/registry"
)

func (s *DeployMonkey) AutodeployerShutdown(ctx context.Context, req *common.Void) (*common.Void, error) {
	ScanAutodeployers()
	PrintAutodeployers()
	return &common.Void{}, nil
}
func (s *DeployMonkey) AutodeployerStartup(ctx context.Context, req *common.Void) (*common.Void, error) {
	fmt.Printf("Autodeployer startedup.\n")
	ScanAutodeployers()
	PrintAutodeployers()
	/*
		sa := &rpb.ServiceAddress{IP:[fromcontext],Port:[fromcontext]}
		err := ScanAutodeployer(sa)
		if err != nil {
			fmt.Printf("Failed to scan autodeployer: %s\n", err)
		}
	*/
	return &common.Void{}, nil
}
