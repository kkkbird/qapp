package main

import (
	"context"

	"github.com/kkkbird/bshark"
	"github.com/kkkbird/bshark/examples/full/pkg/db"
	"github.com/kkkbird/bshark/examples/full/pkg/httpsrv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	appName = "app1"
	addr    = ":8080"
)

func preload() error {
	pflag.String("token", "app1", "appname")
	pflag.String("addr", ":8080", "listen address")

	//viper.Set("file", "app.yml")
	return nil
}

func initDB(ctx context.Context) (bshark.CleanFunc, error) {
	db.Init(ctx, appName)

	return nil, nil
}

func runHTTPServer(ctx context.Context) error {
	return httpsrv.Run(ctx, viper.GetString("token"), viper.GetString("addr"))
}

func onConfigChange() {
	httpsrv.Restart(viper.GetString("token"), viper.GetString("addr"))
}

func main() {
	bshark.New(appName, bshark.WithPreload(preload), bshark.WithConfigChanged(onConfigChange)).
		AddInitStage("initDB", initDB).
		AddDaemons(runHTTPServer).
		Run()
}
