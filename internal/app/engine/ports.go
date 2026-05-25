package engine

import (
	"context"
	"time"

	"github.com/exanubes/appsync/internal/app"
)

type StartEngineInput struct {
	Timeout time.Duration
}

type IO interface {
	Reader
	Writer
}

type Reader interface {
	Read(context.Context) error
}

type Writer interface {
	Write(context.Context) error
}

type Runtime interface {
	Run(context.Context) error
}

type Inbox interface {
	Next(context.Context) (app.Message, error)
}

type ConnectionState interface {
	Close(error)
}
