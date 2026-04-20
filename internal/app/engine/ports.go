package engine

import (
	"context"
	"time"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/queue"
)

type StartEngineInput struct {
	Timeout time.Duration
	Ingress *queue.IngressQueue
	Egress  *queue.EgressQueue
}

type IO interface {
	Reader
	Writer
}

type Reader interface {
	Read(context.Context, *queue.IngressQueue) error
}

type Writer interface {
	Write(context.Context, *queue.EgressQueue) error
}

type Runtime interface {
	Run(context.Context, app.Inbox) error
}
