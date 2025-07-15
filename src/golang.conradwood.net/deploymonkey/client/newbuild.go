package main

import (
	"fmt"

	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/go-easyops/utils"
)

func newbuild() error {
	buildid := 1
	artefactid := 1
	repoid := 1
	commitid := "foocommit"
	def_domain := "conradwood.net"
	yaml, err := utils.ReadFile(*filename)
	if err != nil {
		return err
	}
	nbar := &pb.NewBuildAvailableRequest{
		DeployYaml:   yaml,
		ArtefactID:   uint64(artefactid),
		BuildRepoID:  def_domain,
		BuildID:      uint64(buildid),
		CommitID:     commitid,
		Branch:       "master",
		RepositoryID: uint64(repoid),
	}
	ctx := Context()
	_, err = depl.NewBuildAvailable(ctx, nbar)
	if err != nil {
		fmt.Printf("Failed to do newbuild: %s\n", utils.ErrorString(err))
		return err
	}
	return nil
}
