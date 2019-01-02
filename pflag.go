package bshark

import (
	"os"

	"github.com/kkkbird/qlog"
	"github.com/spf13/pflag"
)

func ParsePFlags(cmdline *pflag.FlagSet) error {
	if cmdline == nil {
		cmdline = pflag.CommandLine
	}
	return cmdline.Parse(qlog.FilterFlags(os.Args[1:]))
}
