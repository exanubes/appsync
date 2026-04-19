package engine

import (
	"context"
	"time"
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
	Run(context.Context, time.Duration) error
}
