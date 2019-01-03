package debugserver

import (
	"context"
	"expvar"
	"html/template"
	"net/http"
	"net/http/pprof"

	"github.com/spf13/viper"

	"github.com/kkkbird/qlog"

	"github.com/spf13/pflag"
)

var (
	debugServeMux *http.ServeMux
	log           = qlog.WithField("bshark", "debugserver")
)

const (
	FlagDebugEnabled = "debugserver.enabled"
	FlagDebugAddr    = "debugserver.addr"
	FlagDebugToken   = "debugserver.token"
)

func init() {
	debugServeMux = http.NewServeMux()
	debugServeMux.HandleFunc("/", debugIndex)
	debugServeMux.HandleFunc("/debug/pprof/", pprof.Index)
	debugServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	debugServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	debugServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	debugServeMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	debugServeMux.Handle("/debug/vars", expvar.Handler())
}

var indexHtml = `
<html>
	<h1>Debug server: {{.Token}}</h1>
	<ul>
		<li><a href="/debug/pprof/">pprof</a></li>
		<li><a href="/debug/vars">var</a></li>
	</ul>
</html>
`

func debugIndex(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("index").Parse(indexHtml))
	t.Execute(w, struct {
		Token string
	}{
		Token: viper.GetString(FlagDebugToken),
	})
}

func AddParam(name string, getter func() interface{}) {
	expvar.Publish(name, expvar.Func(getter))
}

func RegisteDebugServerPFlags() error {
	pflag.Bool(FlagDebugEnabled, false, "enable debug server")
	pflag.String(FlagDebugAddr, ":15050", "listen address")
	pflag.String(FlagDebugToken, "app", "listen address")

	return nil
}

func Run(ctx context.Context) error {
	enabled := viper.GetBool(FlagDebugEnabled)
	token := viper.GetString(FlagDebugToken)
	addr := viper.GetString(FlagDebugAddr)

	if !enabled {
		log.Infof("Debug server is not enabled")
		return nil
	}

	AddParam("token", func() interface{} { return token })

	log.Infof("Debug server for %s start at %s", token, addr)
	return http.ListenAndServe(addr, debugServeMux)
}
