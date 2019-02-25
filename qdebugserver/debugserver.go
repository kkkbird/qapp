package qdebugserver

import (
	"context"
	"expvar"
	"html/template"
	"net/http"

	"github.com/kkkbird/bshark/qhttp"

	"github.com/spf13/viper"

	"github.com/kkkbird/qlog"

	"github.com/spf13/pflag"
)

var (
	debugServeMux = http.NewServeMux()
	log           = qlog.WithField("bshark", "debugserver")
)

const (
	FlagDebugEnabled = "debugserver.enabled"
	FlagDebugAddr    = "debugserver.addr"
)

var indexHTML = `
<html>
	<h1>Debug server</h1>
	<ul>
		<li><a href="{{.Prefix}}pprof">pprof</a></li>
		<li><a href="{{.Prefix}}vars">vars</a></li>
	</ul>
</html>
`

func debugIndex(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("index").Parse(indexHTML))

	t.Execute(w, map[string]interface{}{
		"Prefix": r.URL.Path,
	})
}

func AddParam(name string, getter func() interface{}) {
	expvar.Publish(name, expvar.Func(getter))
}

func RegisteDebugServerPFlags() error {
	pflag.Bool(FlagDebugEnabled, false, "enable debug server")
	pflag.String(FlagDebugAddr, ":15050", "listen address")

	return nil
}

func Run(ctx context.Context) error {
	enabled := viper.GetBool(FlagDebugEnabled)
	addr := viper.GetString(FlagDebugAddr)

	if !enabled {
		log.Infof("Debug server is not enabled")
		return nil
	}

	log.Infof("Debug server start at %s", addr)
	//return http.ListenAndServe(addr, RegisterHTTPMux(debugServeMux))
	return qhttp.RunServer(ctx, addr, RegisterHTTPMux(debugServeMux))
}
