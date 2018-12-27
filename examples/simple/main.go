package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kkkbird/bshark"
	log "github.com/kkkbird/qlog"
)

func initDBSimple(ctx context.Context) (context.Context, error) {
	log.Debugf("Call initDBSimple()")
	return ctx, nil
}

func initDBSimpleFail(ctx context.Context) (context.Context, error) {
	err := fmt.Errorf("Call initDBSimpleFail()")
	log.Error(err)
	return ctx, err
}

func initDBDummy(dummyID int) bshark.InitFunc {
	return func(ctx context.Context) (context.Context, error) {
		log.Debugf("Call initDBDummy():%d", dummyID)
		return ctx, nil
	}
}

type cKey int

const (
	cHTTPName cKey = iota
	cHTTPPort
)

func initHTTPServer(ctx context.Context) (context.Context, error) {
	ctx = context.WithValue(ctx, cHTTPName, "simplehttp")
	ctx = context.WithValue(ctx, cHTTPPort, ":8080")

	return ctx, nil
}

func indexHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world from %s", name)
	}
}

func runHTTPServerSimple(ctx context.Context) error {
	name := ctx.Value(cHTTPName).(string)
	port := ctx.Value(cHTTPPort).(string)

	srv := http.NewServeMux()
	srv.HandleFunc("/", indexHandler(name))

	log.WithField("name", name).WithField("port", port).Debugf("Run runHTTPServerSimple()")

	server := &http.Server{
		Addr:    port,
		Handler: srv,
	}
	return server.ListenAndServe()
}

func runHTTPServerDummy(port string) bshark.DaemonFunc {

	return func(ctx context.Context) error {
		srv := http.NewServeMux()
		srv.HandleFunc("/", indexHandler(port))

		log.WithField("name", port).WithField("port", port).Debugf("Run runHTTPServerDummy()")

		server := &http.Server{
			Addr:    port,
			Handler: srv,
		}
		return server.ListenAndServe()
	}
}

func main() {
	bshark.New("mytestapp", nil).SetLogger(log.WithField("bshark", "mytestapp")).
		AddInitStage("initDB", initDBSimple).
		AddInitStage("initDBs", initDBDummy(2), initDBDummy(3), initDBDummy(4)).
		AddInitStage("initHTTPServer", initHTTPServer).
		AddDaemons(runHTTPServerSimple, runHTTPServerDummy(":18080")).
		AddDaemons(runHTTPServerDummy(":18081"), runHTTPServerDummy(":18082")).
		Run()
}
