package main

import (
	pb "golang.conradwood.net/apis/deploymonkey"
	"golang.conradwood.net/deploymonkey/db"
)

func save_app_limits(appid uint64, limits *pb.Limits) error {
	sl := &pb.SQLLimits{
		AppDef:        &pb.ApplicationDefinition{ID: appid},
		MaxMemory:     limits.MaxMemory,
		Priority:      limits.Priority,
		MaxKillMemory: limits.MaxKillMemory,
		MaxSwapMemory: limits.MaxSwapMemory,
	}
	_, err := db.DefaultDBSQLLimits().Save(TEMPCONTEXT(), sl)

	//	_, err := dbcon.ExecContext(TEMPCONTEXT(), "saveapplimits", "INSERT into applimits (app_id,maxmemory,priority) values ($1,$2,$3)", appid, limits.MaxMemory, limits.Priority)
	return err
}

// given an application id, loads "app limits"
func AppLimitsByAppID(appid uint64) (*pb.Limits, error) {
	arows, err := db.DefaultDBSQLLimits().ByAppDef(TEMPCONTEXT(), appid)
	if err != nil {
		return nil, err
	}
	if len(arows) == 0 {
		return nil, nil
	}
	limits := arows[0]
	l := &pb.Limits{
		MaxMemory:     limits.MaxMemory,
		Priority:      limits.Priority,
		MaxKillMemory: limits.MaxKillMemory,
		MaxSwapMemory: limits.MaxSwapMemory,
	}
	return l, nil
	// old version:
	/*
		// add new limits here...
		rows, err := dbcon.QueryContext(TEMPCONTEXT(), "loadapplimits", "SELECT maxmemory,priority from applimits where app_id = $1", appid)
		if err != nil {
			fmt.Printf("Failed to get app with id %d:%s\n", appid, err)
			return nil, err
		}
		defer rows.Close()
		if !rows.Next() {
			return nil, nil
		}
		res := &pb.Limits{}
		err = rows.Scan(&res.MaxMemory, &res.Priority)
		if err != nil {
			return nil, err
		}

		return res, nil
	*/
}
