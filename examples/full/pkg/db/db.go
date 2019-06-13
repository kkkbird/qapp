package db

import (
	"context"

	log "github.com/sirupsen/logrus"
)

// InitDB is an example
func Init(ctx context.Context, name string) error {

	log.Debugf("call initdb by %s", name)

	return nil
}
