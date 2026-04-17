package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/exanubes/appsync/internal/app"
)

type Slogger struct {
	context string
	logger  *slog.Logger
}

func create_logger(context string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	if context != "" {
		return logger.With("context", context)
	}

	return logger
}

func New() Slogger {
	return Slogger{
		logger: create_logger(""),
	}
}

func (l Slogger) Debug(message string, params ...any) {
	var builder strings.Builder
	builder.WriteString(message)
	l.logger.Debug(builder.String(), params...)
}

func (logger Slogger) SetContext(context string) app.Logger {
	logger.context = context

	// logger.logger = logger.logger.With("context", context)
	logger.logger = create_logger(context)
	return logger
}
