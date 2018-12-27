package bshark

// Logger is a simple interface for bshark textout messages
type Logger interface {
	Printf(format string, args ...interface{})
}
