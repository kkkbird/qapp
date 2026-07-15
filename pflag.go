package qapp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/kkkbird/qapp/qdebugserver"
	"github.com/kkkbird/qlog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Predefined errors
var (
	ErrShowVersion = errors.New("ErrShowVersion")
)

func (a *Application) handleFlagsAndEnv() error {
	var err error
	qdebugserver.RegisteDebugServerPFlags()

	pflag.StringP("file", "f", "app.yml", "config file name")
	pflag.BoolP("version", "v", false, "show version")

	if a.preload != nil {
		if err = a.preload(); err != nil {
			log.WithError(err).Debug("preload error")
			return err
		}
	}

	a.cmdline.Parse(qlog.FilterFlags(os.Args[1:]))

	// bind pflags
	viper.BindPFlags(pflag.CommandLine)

	// bind env
	viper.AutomaticEnv()
	if len(a.envPrefix) > 0 {
		viper.SetEnvPrefix(a.envPrefix)
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// if just show version
	if viper.GetBool("version") {
		showAppVersion(os.Stdout, a.name)
		return ErrShowVersion
	}

	// read from config file
	configFile := viper.GetString("file")
	viper.SetConfigFile(configFile)
	err = viper.ReadInConfig() // Find and read the config file

	if err != nil { // Handle errors reading the config file
		log.WithError(err).Debug("Read from config fail, use default settings") // it is ok that we cannot read from config file
	} else {
		// merge override config file if it exists, values in it will override the main config
		if mergeErr := a.mergeOverrideConfig(configFile); mergeErr != nil {
			log.WithError(mergeErr).Debug("Read from override config fail, skipped")
		}

		// watch config change
		if a.onConfigFileChanged != nil {
			viper.WatchConfig()
			viper.OnConfigChange(func(e fsnotify.Event) {
				log.Trace("Config file changed:", e.Name)
				a.onConfigFileChanged()
			})
		}
	}
	return nil
}

// mergeOverrideConfig merges an override config file (if it exists) on top of the
// already-loaded main config. Values present in the override file take precedence.
// The override file name is resolved via WithConfigOverrideFile when set, otherwise
// it is derived from the main config file name, e.g. "app.yml" -> "app.override.yml".
func (a *Application) mergeOverrideConfig(mainConfigFile string) error {
	overrideFile := a.configOverrideFile
	if overrideFile == "" {
		overrideFile = deriveOverrideConfigFile(mainConfigFile)
	}

	// skip silently when the override file does not exist
	if _, err := os.Stat(overrideFile); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	log.Infof("Merging override config: %s", overrideFile)

	viper.SetConfigFile(overrideFile)
	return viper.MergeInConfig()
}

// deriveOverrideConfigFile builds an override config file name from the main one by
// inserting ".override" before the extension, e.g. "app.yml" -> "app.override.yml".
// Files without an extension get an ".override" suffix, e.g. "app" -> "app.override".
func deriveOverrideConfigFile(file string) string {
	dir := filepath.Dir(file)
	name := filepath.Base(file)

	if dot := strings.LastIndex(name, "."); dot > 0 {
		base := name[:dot]
		ext := name[dot:]
		return filepath.Join(dir, fmt.Sprintf("%s.override%s", base, ext))
	}
	return filepath.Join(dir, fmt.Sprintf("%s.override", name))
}

// WithCmdLine set init with a timeout
func WithCmdLine(cmdline *pflag.FlagSet) AppOpts {
	return func(a *Application) {
		a.cmdline = cmdline
	}
}

// WithConfigChanged set config change handler for app
func WithConfigChanged(onConfigChange func()) AppOpts {
	return func(a *Application) {
		a.onConfigFileChanged = onConfigChange
	}
}

// WithConfigOverrideFile sets an explicit override config file name. When this file
// exists at startup, its values are merged on top of the main config file (specified
// by -f/--file) and override matching keys. If not set, it defaults to a file named
// "<main>.override.<ext>" alongside the main config (e.g. "app.override.yml").
func WithConfigOverrideFile(name string) AppOpts {
	return func(a *Application) {
		a.configOverrideFile = name
	}
}

// WithPreload set preload func,
// preload is only used to set flags or some common setting, normally, we should use init func
func WithPreload(preload func() error) AppOpts {
	return func(a *Application) {
		a.preload = preload
	}
}

// WithEnvPrefix set env prefix, ex WithEnvPrefix("qapp"), envs used by qapp should be prefixed by "QAPP_"
func WithEnvPrefix(envPrefix string) AppOpts {
	return func(a *Application) {
		a.envPrefix = envPrefix
	}
}
