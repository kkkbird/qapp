package bshark

import (
	"os"
	"text/template"
)

// predefined version params
var (
	Version   = "unknown-version"
	BuildTime = "unknown-buildtime"
	GitHash   = "unknown-githash"
	GoVersion = "unknown-goversion"
)

var versionTemplate = `
App: {{.Name}}
  Version:      {{.Version}}
  Build time:   {{.BuildTime}}
  GitHash:      {{.GitHash}}
  Go version:   {{.GoVersion}}
`

func showAppVersion(name string) error {
	t := template.Must(template.New("version").Parse(versionTemplate))

	err := t.Execute(os.Stdout, map[string]string{
		"Name":      name,
		"Version":   Version,
		"BuildTime": BuildTime,
		"GitHash":   GitHash,
		"GoVersion": GoVersion,
	})

	return err
}
