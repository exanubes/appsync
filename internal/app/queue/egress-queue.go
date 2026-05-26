package queue

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type EgressQueue struct {
	inbox      chan []byte
	connection ConnectionState
}

func NewEgressQueue(max_size uint, connection ConnectionState) *EgressQueue {
	return &EgressQueue{
		inbox:      make(chan []byte, max_size),
		connection: connection,
	}
}

func (registry *EgressQueue) Next(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case payload := <-registry.inbox:
		return payload, nil
	}
}

func (registry *EgressQueue) Enqueue(ctx context.Context, payload []byte) error {
	select {
	case registry.inbox <- payload:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-registry.connection.Done():
		return app.ErrConnectionClosed
	}
}
