package db

import "context"

/*
		a custom column handler can be attached to a dbquerier to manage columns which are not part of the proto
	        Typical usecases:
	        - maintain a "lastupdatedby" (userid) column
	        - maintain a "lastupdatedat" (timestamp) column
	        - maintain a "owner" (userid) column, which may be used to limit queries as well
*/
type CustomColumnHandler interface {
	// return map of column names and values to query for (returning nil,nil is supported)
	FieldsToQuery(ctx context.Context) (map[string]interface{}, error)
	// return map of column names and values to store in database in addition to the proto (returning nil,nil is supported)
	FieldsToStore(ctx context.Context, p interface{}) (map[string]interface{}, error)
}

type DBQuerier interface {
	GetCustomColumnHandlers() []CustomColumnHandler
}

func extraFieldsToStore(ctx context.Context, a DBQuerier, p interface{}) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	for _, c := range a.GetCustomColumnHandlers() {
		ares, err := c.FieldsToStore(ctx, p)
		if err != nil {
			return nil, err
		}
		if ares == nil {
			continue
		}
		for k, v := range ares {
			res[k] = v
		}
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res, nil
}

func extraFieldsToQuery(ctx context.Context, a DBQuerier) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	for _, c := range a.GetCustomColumnHandlers() {
		ares, err := c.FieldsToQuery(ctx)
		if err != nil {
			return nil, err
		}
		if ares == nil {
			continue
		}
		for k, v := range ares {
			res[k] = v
		}
	}
	return res, nil
}
