package main

import (
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/deployminator"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strconv"
)

var (
	repoid    = flag.Uint64("repository_id", 0, "the repository id of this build")
	buildid   = flag.Uint64("build_id", 0, "the  id of this build")
	do_import = flag.Bool("import", false, "import build repo latest builds")
)

func main() {
	flag.Parse()
	if *do_import {
		DoImport()
		os.Exit(0)
	}
	filename := ""
	if len(flag.Args()) == 0 {
		fmt.Printf("Please supply a filename\n")
		os.Exit(0)
	}
	filename = flag.Args()[0]
	b, err := utils.ReadFile(filename)
	utils.Bail("failed to read file", err)
	ctx := authremote.Context()
	br := &pb.NewBuildRequest{
		DeployFile:   b,
		RepositoryID: *repoid,
		Branch:       "master",
		BuildNumber:  *buildid,
	}
	_, err = pb.GetDeployminatorClient().NewBuildAvailable(ctx, br)
	utils.Bail("failed to submit new build", err)
	fmt.Printf("Done\n")
}

func DoImport() {
	dirs, err := ioutil.ReadDir("/srv/build-repository/artefacts")
	utils.Bail("failed to read artefacts", err)
	for _, d := range dirs {
		doArtefact(d.Name())
	}
}
func doArtefact(name string) {
	latest_version := findLatestArtefact(name)
	if latest_version == 0 {
		return
	}
	repoid := findRepoID(name, latest_version)
	fmt.Printf("Artefact: \"%s\": %d (repo %d)\n", name, latest_version, repoid)
	if repoid == 0 {
		return
	}
	fname := fmt.Sprintf("/srv/build-repository/artefacts/%s/master/%d/deployment/deploy.yaml", name, latest_version)
	if !utils.FileExists(fname) {
		return
	}
	b, err := utils.ReadFile(fname)
	utils.Bail("failed to read deploy file", err)
	ctx := authremote.Context()
	br := &pb.NewBuildRequest{
		DeployFile:   b,
		RepositoryID: repoid,
		Branch:       "master",
		BuildNumber:  latest_version,
	}
	_, err = pb.GetDeployminatorClient().NewBuildAvailable(ctx, br)
	utils.Bail("failed to add build", err)
}

func findLatestArtefact(name string) uint64 {
	metadir, err := ioutil.ReadDir(fmt.Sprintf("/srv/build-repository/metadata/%s/master", name))
	utils.Bail("no metadata", err)
	latest := uint64(0)
	for _, md := range metadir {
		n := md.Name()
		vs, err := strconv.Atoi(n)
		if err != nil {
			fmt.Printf("not a version: %s/%s", name, err)
			continue
		}
		v := uint64(vs)
		if v > latest {
			latest = v
		}
	}
	return latest
}

type builddef struct {
	Buildid      int
	Repositoryid uint64
}

func findRepoID(name string, version uint64) uint64 {
	bd := readMeta(name, version)
	return bd.Repositoryid
}
func readMeta(name string, version uint64) *builddef {
	fname := fmt.Sprintf("/srv/build-repository/metadata/%s/master/%d/build.yaml", name, version)
	rb, err := utils.ReadFile(fname)
	utils.Bail("no metadata", err)
	bd := &builddef{}
	err = yaml.Unmarshal(rb, bd)
	utils.Bail(fmt.Sprintf("invalid yaml in %s", fname), err)
	return bd
}
