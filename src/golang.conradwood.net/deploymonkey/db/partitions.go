package db

import (
	"context"

	"golang.conradwood.net/go-easyops/authremote"
)

type partitionHandler struct {
	col string
}

func NewPartitionHandler(colname string) CustomColumnHandler {
	return &partitionHandler{col: colname}
}

// return map of column names and values to query for
func (cch *partitionHandler) FieldsToQuery(ctx context.Context) (map[string]interface{}, error) {
	pid, err := authremote.PartitionID(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{cch.col: pid}, nil
}

// return map of column names and values to store in database in addition to the proto
func (cch *partitionHandler) FieldsToStore(ctx context.Context, p interface{}) (map[string]interface{}, error) {
	pid, err := authremote.PartitionID(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{cch.col: pid}, nil
}
