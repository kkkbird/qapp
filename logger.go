package bshark

type Logger interface {
	Printf(format string, args ...interface{})
}
