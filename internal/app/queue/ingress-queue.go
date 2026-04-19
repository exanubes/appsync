package queue

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type IngressQueue struct {
	inbox chan app.Message
}

func NewIngressQueue(max_size uint) *IngressQueue {
	return &IngressQueue{
		inbox: make(chan app.Message, max_size),
	}
}

func (registry *IngressQueue) Next() {}

func (registry *IngressQueue) Enqueue(ctx context.Context, msg app.Message) error {
	select {
	case registry.inbox <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
