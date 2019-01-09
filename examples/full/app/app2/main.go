package main

import (
	"context"

	"github.com/kkkbird/bshark/debugserver"

	"github.com/gin-gonic/gin"
	"github.com/kkkbird/bshark"
	"github.com/kkkbird/bshark/examples/full/pkg/db"
)

const (
	appName = "app2"
	addr    = ":8082"
)

func initDB(ctx context.Context) (bshark.CleanFunc, error) {
	db.Init(ctx, appName)

	return nil, nil
}

func runHTTPServer(ctx context.Context) error {
	r := gin.Default()
	debugserver.RegisterGin(r, "/dev")

	return r.Run()
}

func main() {
	bshark.New(appName).
		AddInitStage("initDB", initDB).
		AddDaemons(runHTTPServer).
		Run()
}
