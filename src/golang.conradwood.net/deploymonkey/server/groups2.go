package main

import (
	pb "golang.conradwood.net/apis/deploymonkey"
)

type Group2Handler struct {
}

func (g *Group2Handler) CreateOrFindGroupVersion([]*pb.ApplicationDefinition) *pb.GroupVersion {
}
