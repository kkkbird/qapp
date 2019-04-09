package main

import (
	"context"

	"github.com/kkkbird/qapp/qhttp"

	"github.com/kkkbird/qapp/qdebugserver"

	"github.com/gin-gonic/gin"
	"github.com/kkkbird/qapp"
	"github.com/kkkbird/qapp/examples/full/pkg/db"
)

const (
	appName = "app2"
	addr    = ":8088"
)

func initDB(ctx context.Context) (qapp.CleanFunc, error) {
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
	qapp.New(appName).
		AddInitStage("initDB", initDB).
		AddDaemons(runHTTPServer).
		Run()
}
