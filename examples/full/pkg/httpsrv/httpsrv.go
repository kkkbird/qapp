package httpsrv

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

var (
	srv *http.Server

	restartChan = make(chan context.Context, 1)
	errChan     = make(chan error, 1)

	log = logrus.WithField("app", "gin")
)

func qGinLogger(c *gin.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery

	// Process request
	c.Next()

	//if _, ok := skip[path]; !ok { // TODO add skip path as gin

	if raw != "" {
		path = path + "?" + raw
	}

	log.WithFields(logrus.Fields{
		"StatusCode": c.Writer.Status(),
		"Method":     c.Request.Method,
		"Latency":    time.Now().Sub(start),
		"ClientIP":   c.ClientIP(),
	}).Debug(path)
}

func startServer(name string, addr string) error {
	r := gin.New()

	// Global middleware
	// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	r.Use(qGinLogger)
	//r.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// Per route middleware, you can add as many as you desire.
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("pong from %s", name))
	})

	srv = &http.Server{
		Addr:    addr,
		Handler: r,
	}
	log.Debugf("Server %s start at %s", name, addr)
	// service connections
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	log.Infof("Server %s stopped", name)

	return nil
}

func shutdownServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

func Restart(name string, addr string) {
	ctx := context.WithValue(context.Background(), "name", name)
	ctx = context.WithValue(ctx, "addr", addr)
	restartChan <- ctx
}

// RunHTTPServer is an example function
func Run(ctx context.Context, name string, addr string) error {
	go func() {
		if err := startServer(name, addr); err != nil {
			errChan <- err
		}
	}()

	var err error

__EndRun:
	for {
		select {
		case <-ctx.Done():
			err = shutdownServer()
			break __EndRun
		case newCtx := <-restartChan:
			name := newCtx.Value("name").(string)
			addr := newCtx.Value("addr").(string)
			if err = shutdownServer(); err != nil {
				break __EndRun
			}
			go func() {
				if err := startServer(name, addr); err != nil {
					errChan <- err
				}
			}()
		case err = <-errChan:
			break __EndRun
		}
	}

	return err
}
