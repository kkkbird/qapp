package bshark

import (
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/kkkbird/bshark/debugserver"
	"github.com/kkkbird/qlog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func (a *Application) handleFlagsAndEnv() {
	debugserver.RegisteDebugServerPFlags()
	pflag.StringP("file", "f", "app.yml", "config file name")

	if a.registerAppFlags != nil {
		a.registerAppFlags()
	}

	a.cmdline.Parse(qlog.FilterFlags(os.Args[1:]))

	// bind pflags
	viper.BindPFlags(pflag.CommandLine)

	// bind env
	viper.AutomaticEnv()

	// read from config file
	viper.SetConfigFile(viper.GetString("file"))
	err := viper.ReadInConfig() // Find and read the config file

	if err != nil { // Handle errors reading the config file
		//return errors.New("read error fail")
	} else {
		// watch config change
		if a.onConfigFileChanged != nil {
			viper.WatchConfig()
			viper.OnConfigChange(func(e fsnotify.Event) {
				log.Debugln("Config file changed:", e.Name)
				a.onConfigFileChanged()
			})
		}
	}
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

// WithAppFlagRegister set app self flags
func WithAppFlagRegister(appFlagRegister func()) AppOpts {
	return func(a *Application) {
		a.registerAppFlags = appFlagRegister
	}
}
