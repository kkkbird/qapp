package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/kkkbird/qapp/qhttp"

	"github.com/kkkbird/qapp"
	log "github.com/kkkbird/qlog"
)

func initDBSimple(ctx context.Context) (qapp.CleanFunc, error) {
	log.Debugf("Call initDBSimple()")
	return nil, nil
}

func initDBSimpleFail(ctx context.Context) (qapp.CleanFunc, error) {
	err := fmt.Errorf("Call initDBSimpleFail()")
	log.Error(err)
	return nil, err
}

func initDBSimpleTimeout(ctx context.Context) (qapp.CleanFunc, error) {
	log.Debug("Call initDBSimpleTimeout() start")
	time.Sleep(5 * time.Second)
	log.Debug("Call initDBSimpleTimeout() end")
	return nil, nil
}

func initDBSimpleTimeoutWithContext(ctx context.Context) (qapp.CleanFunc, error) {
	log.Debug("Call initDBSimpleTimeoutWithContext() start")
	select {
	case <-time.After(5 * time.Second):
		log.Debug("Call initDBSimpleTimeoutWithContext() done")
	case <-ctx.Done():
		log.Debug("Finish initDBSimpleTimeoutWithContext() by context")
	}
	log.Debug("Call initDBSimpleTimeout() end")
	return nil, nil
}

func initDBDummy(dummyID int) qapp.InitFunc {
	return func(ctx context.Context) (qapp.CleanFunc, error) {
		log.Debugf("Call initDBDummy():%d", dummyID)
		return nil, nil
	}
}

type cKey int

const (
	cHTTPName cKey = iota
	cHTTPPort
)

func initHTTPServer(ctx context.Context) (qapp.CleanFunc, error) {
	ctx = context.WithValue(ctx, cHTTPName, "simplehttp")
	ctx = context.WithValue(ctx, cHTTPPort, ":8080")

	return nil, nil
}

func indexHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world from %s", name)
	}
}

func runHTTPServerSimple(ctx context.Context) error {
	// name := ctx.Value(cHTTPName).(string)
	// port := ctx.Value(cHTTPPort).(string)
	name := "simpleServer"
	port := ":8080"

	srv := http.NewServeMux()
	srv.HandleFunc("/", indexHandler(name))

	log.WithField("name", name).WithField("port", port).Debugf("Run runHTTPServerSimple()")

	// server := &http.Server{
	// 	Addr:    port,
	// 	Handler: srv,
	// }
	// return server.ListenAndServe()

	return qhttp.RunServer(ctx, port, srv)
}

func runDaemonFail(ctx context.Context) error {
	return errors.New("error runHTTPServerSimple2")
}

func runHTTPServerDummy(port string) qapp.DaemonFunc {
	return func(ctx context.Context) error {
		srv := http.NewServeMux()
		srv.HandleFunc("/", indexHandler(port))

		log.WithField("name", port).WithField("port", port).Debugf("Run runHTTPServerDummy()")

		// server := &http.Server{
		// 	Addr:    port,
		// 	Handler: srv,
		// }

		// return server.ListenAndServe()
		return qhttp.RunServer(ctx, port, srv)
	}
}

func preload() error {
	return nil
}

func main() {
	qapp.New("mytestapp", qapp.WithInitTimeout(3*time.Second), qapp.WithPreload(preload)).
		AddInitStage("initDB", initDBSimple).
		AddInitStage("initDBs", initDBDummy(2), initDBDummy(3), initDBDummy(4)).
		//AddInitStage("initDbs2", initDBSimpleTimeout, initDBSimpleTimeoutWithContext, initDBSimpleFail).
		AddInitStage("initHTTPServer", initHTTPServer).
		AddDaemons(runHTTPServerDummy(":18080")).
		AddDaemons(runHTTPServerDummy(":18081"), runHTTPServerDummy(":18082")).
		//AddDaemons(runDaemonFail).
		Run()
}
