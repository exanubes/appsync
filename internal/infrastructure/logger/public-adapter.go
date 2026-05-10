package logger

import (
	"github.com/exanubes/appsync/internal/app"
	appsynclogger "github.com/exanubes/appsync/logger"
)

type InternalLoggerAdapter struct {
	public appsynclogger.Logger
}

func NewInternalLoggerAdapter(public appsynclogger.Logger) *InternalLoggerAdapter {
	return &InternalLoggerAdapter{public: public}
}

func (a *InternalLoggerAdapter) Debug(message string, params ...any) {
	a.public.Debug(message, params...)
}

func (a *InternalLoggerAdapter) SetContext(context string) app.Logger {
	return NewInternalLoggerAdapter(a.public.SetContext(context))
}
