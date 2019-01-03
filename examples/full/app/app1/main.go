package main

import (
	"context"

	"github.com/kkkbird/bshark"
	"github.com/kkkbird/bshark/debugserver"
	"github.com/kkkbird/bshark/examples/full/pkg/db"
	"github.com/kkkbird/bshark/examples/full/pkg/httpsrv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	appName = "app1"
	addr    = ":8080"
)

func regFlags() {
	pflag.String("token", "app1", "appname")
	pflag.String("addr", ":8080", "listen address")

	viper.RegisterAlias(debugserver.FlagDebugToken, "token")
	viper.Set("file", "app.yml")
}

func initDB(ctx context.Context) error {
	db.InitDB(ctx, appName)

	return nil
}

func runHTTPServer(ctx context.Context) error {
	return httpsrv.RunHTTPServer(ctx, viper.GetString("token"), viper.GetString("addr"))
}

func main() {
	bshark.New(appName, bshark.WithPreInit(regFlags)).
		AddInitStage("initDB", initDB).
		AddDaemons(runHTTPServer).
		Run()
}
