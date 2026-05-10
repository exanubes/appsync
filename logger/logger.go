package logger

import (
	"log/slog"
	"os"
)

// Logger is the interface for debug logging within the library.
// Implement this to plug in a custom logger; or use [New] for a slog-based default.
type Logger interface {
	Debug(string, ...any)
	SetContext(string) Logger
}

type slogger struct {
	l *slog.Logger
}

// New returns a Logger that writes JSON-formatted debug output to stdout.
func New() Logger {
	return newSlogger("")
}

func newSlogger(context string) *slogger {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	l := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	if context != "" {
		l = l.With("context", context)
	}
	return &slogger{l: l}
}

func (s *slogger) Debug(message string, params ...any) {
	s.l.Debug(message, params...)
}

func (s *slogger) SetContext(context string) Logger {
	return newSlogger(context)
}
