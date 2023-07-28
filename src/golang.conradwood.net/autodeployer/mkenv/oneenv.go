package mkenv

import (
	"context"
	pb "golang.conradwood.net/apis/commondeploy"
	"golang.conradwood.net/autodeployer/fscache"
)

type oneenv struct {
	req     *pb.MkenvRequest
	ctx     context.Context
	fscache fscache.FSCache
}
