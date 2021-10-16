package main

import (
	"fmt"
	pb "golang.conradwood.net/apis/deploymonkey"
)

func AppToString(app *pb.ApplicationDefinition) string {
	return fmt.Sprintf("Build #%d in %d(%s)", app.BuildID, app.RepositoryID, app.Binary)
}
