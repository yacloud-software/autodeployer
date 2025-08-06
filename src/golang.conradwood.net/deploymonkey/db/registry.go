package db

import (
	"context"
	"sync"
)

/*
   convoluted - we get a handler-factory (yes, java style) because during init we do not want to open databases just yet, so we defer until someone
   actually requests the handler
*/

var (
	handler_lock   sync.Mutex
	handlers       []func() Handler
	tables_created = false
)

type Handler interface {
	CreateTable(ctx context.Context) error
}

func RegisterDBHandlerFactory(f func() Handler) {
	handler_lock.Lock()
	defer handler_lock.Unlock()
	handlers = append(handlers, f)
}
func CreateAllTables(ctx context.Context) error {
	if tables_created {
		return nil
	}
	handler_lock.Lock()
	defer handler_lock.Unlock()
	if tables_created {
		return nil
	}
	tables_created = true
	var xerr error
	for _, f := range handlers {
		err := f().CreateTable(ctx)
		if err != nil {
			xerr = err
		}
	}
	return xerr
}
