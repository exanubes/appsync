package logger

import "github.com/exanubes/appsync/internal/app"

type NoopLogger struct {
}

func (NoopLogger) Debug(message string, params ...any) {
}

func (l NoopLogger) SetContext(context string) app.Logger {
	return l
}
