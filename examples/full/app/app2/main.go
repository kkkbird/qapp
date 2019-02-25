package main

import (
	"context"

	"github.com/kkkbird/bshark/qhttp"

	"github.com/kkkbird/bshark/qdebugserver"

	"github.com/gin-gonic/gin"
	"github.com/kkkbird/bshark"
	"github.com/kkkbird/bshark/examples/full/pkg/db"
)

const (
	appName = "app2"
	addr    = ":8088"
)

func initDB(ctx context.Context) (bshark.CleanFunc, error) {
	db.Init(ctx, appName)

	return nil, nil
}

func runHTTPServer(ctx context.Context) error {
	r := gin.Default()
	qdebugserver.RegisterGin(r, "/dev")

	//return r.Run(addr)

	return qhttp.RunServer(ctx, addr, r)
}

func main() {
	bshark.New(appName).
		AddInitStage("initDB", initDB).
		AddDaemons(runHTTPServer).
		Run()
}
