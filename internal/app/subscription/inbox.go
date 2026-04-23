package subscription

import (
	"context"
	"time"

	"github.com/exanubes/appsync/internal/app"
)

const default_enqueue_timeout = 300 * time.Millisecond

type inbox struct {
	queue   chan app.Payload
	timeout time.Duration
	done    <-chan struct{}
}

func new_inbox(done <-chan struct{}, buffer_size uint) *inbox {
	return &inbox{
		queue:   make(chan app.Payload, buffer_size),
		timeout: default_enqueue_timeout,
		done:    done,
	}
}

func (inbox *inbox) Enqueue(ctx context.Context, payload app.Payload) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	timeout := time.After(inbox.timeout)
	select {
	case <-inbox.done:
		return app.ErrSubscriptionClosed
	case <-ctx.Done():
		return ctx.Err()
	case inbox.queue <- payload:
		return nil
	case <-timeout: // INFO: queue is most likely full, for now drop message
		return app.ErrSubscriptionInboxFull
	}
}
func (inbox *inbox) Next(ctx context.Context) (app.Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case <-inbox.done:
		return nil, app.ErrSubscriptionClosed
	case <-ctx.Done():
		return nil, ctx.Err()

	case payload := <-inbox.queue:
		return payload, nil
	}
}
