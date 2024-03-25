package main

import (
	"fmt"
	"strconv"
	"strings"
)

// groupid,buildid,version...
// version may be -1 if not part of this id
func DecodeDeploymentID(deplid string) (int, int, int) {
	if strings.HasPrefix(deplid, "DM-APPDEF2-") {
		return DecodeDeploymentIDv2(deplid)
	}
	if !strings.HasPrefix(deplid, DEPLOY_PREFIX) {
		fmt.Printf("Not a valid deploy_prefix: \"%s\"\n", deplid)
		return -1, -1, -1
	}
	x := deplid[len(DEPLOY_PREFIX)+1:]
	z := strings.Split(x, "-")
	if (len(z) != 2) && (len(z) != 3) {
		fmt.Printf("Not a valid deploy_id: %s\n", deplid)
		return -1, -1, -1
	}
	gid, err := strconv.Atoi(z[0])
	if err != nil {
		fmt.Printf("group invalid deplid %s: %s\n", deplid, err)
		return -1, -1, -1
	}
	bid, err := strconv.Atoi(z[1])
	if err != nil {
		fmt.Printf("build invalid deplid %s: %s\n", deplid, err)
		return -1, -1, -1
	}
	vid := 0
	if len(z) > 2 {
		vid, err = strconv.Atoi(z[2])
		if err != nil {
			fmt.Printf("version invalid deplid %s: %s\n", deplid, err)
			return -1, -1, -1
		}
	}
	return gid, bid, vid

}

func DecodeDeploymentIDv2(deplid string) (int, int, int) {
	fmt.Printf("Unable to parse v2 atm\n")
	return 0, 0, 0
}
