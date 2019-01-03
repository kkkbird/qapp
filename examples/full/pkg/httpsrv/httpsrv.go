package httpsrv

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/kkkbird/qlog"
)

func indexHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world from %s", name)
	}
}

// RunHTTPServer is an example function
func RunHTTPServer(ctx context.Context, name string, addr string) error {
	srv := http.NewServeMux()
	srv.HandleFunc("/", indexHandler(name))

	log.WithField("name", name).WithField("addr", addr).Debugf("Run RunHTTPServer()")

	server := &http.Server{
		Addr:    addr,
		Handler: srv,
	}
	return server.ListenAndServe()
}
