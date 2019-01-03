package db

import (
	"context"

	log "github.com/kkkbird/qlog"
)

// InitDB is an example
func InitDB(ctx context.Context, name string) error {

	log.Debugf("call initdb by %s", name)

	return nil
}
