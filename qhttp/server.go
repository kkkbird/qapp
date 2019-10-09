package qhttp

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	log = logrus.WithField("pkg", "qhttp")
)

// GrpcHandlerFunc bind grpc server and http server to same port, but it not work without tls enabled
// refer to https://github.com/dhrp/grpc-rest-go-example
func GrpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(tamird): point to merged gRPC code rather than a PR.
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

// RunGRPCServer run a GRPC server with gracefule stop
// example: 
// func runGRPCServer(ctx context.Context) error {	
//	 s := grpc.NewServer()
// 	 pb.RegisterGreeterServer(s, &server{})
// 	 return qhttp.RunGRPCServer(ctx, port, s)
// }
func RunGRPCServer(ctx context.Context, addr string, s *grpc.Server) (err error) {
	srvErrChan := make(chan error)

	go func() {
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.WithError(err).Error("GRPC server listen fail")
		}

		srvErrChan <- s.Serve(lis)
	}()

	select {
	case err = <-srvErrChan:
		return err
	case <-ctx.Done():
		shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		go func() {
			s.GracefulStop()
			log.Debug("GRPC server gracefully exit")
			cancel()
		}()

		<-shutDownCtx.Done()
		return nil
	}
}

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
