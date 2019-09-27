package qhttp

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("pkg", "qhttp")
)

// RunServer run a http server with gracefully shutdown
func RunServer(ctx context.Context, addr string, handler http.Handler, opts ...func(*http.Server)) (err error) {
	log.Debugf("Listening and serving HTTP on %s", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	if opts != nil {
		for _, optFunc := range opts {
			optFunc(srv)
		}
	}

	srvErrChan := make(chan error)

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Error("HTTP server error")
			srvErrChan <- err
		}
	}()

	select {
	case err = <-srvErrChan:
		return err
	case <-ctx.Done():
		shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err = srv.Shutdown(shutDownCtx); err != nil {
			log.WithError(err).Error("HTTP server Shutdown:", err)
			return err
		}
	}

	log.Debug("HTTP server gracefully exit")
	return nil
}
