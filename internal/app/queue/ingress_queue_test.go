package queue_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/queue"
)

func TestIngressQueue_Next(t *testing.T) {
	var msg app.Message = "test-message"

	tests := []struct {
		name       string
		setup      func(*queue.IngressQueue)
		ctx        func() context.Context
		expect_msg app.Message
		expect_err error
	}{
		{
			name: "returns message when available",
			setup: func(q *queue.IngressQueue) {
				q.Enqueue(context.Background(), msg)
			},
			ctx:        context.Background,
			expect_msg: msg,
			expect_err: nil,
		},
		{
			name:  "returns context error when context is cancelled",
			setup: func(_ *queue.IngressQueue) {},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expect_msg: nil,
			expect_err: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := queue.NewIngressQueue(1)
			tt.setup(q)

			got, err := q.Next(tt.ctx())

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("Next() error = %v, want %v", err, tt.expect_err)
			}
			if got != tt.expect_msg {
				t.Errorf("Next() = %v, want %v", got, tt.expect_msg)
			}
		})
	}
}

func TestIngressQueue_Enqueue(t *testing.T) {
	var msg app.Message = "test-message"

	tests := []struct {
		name       string
		setup      func(*queue.IngressQueue)
		ctx        func() context.Context
		msg        app.Message
		expect_err error
	}{
		{
			name:       "enqueues message successfully",
			setup:      func(_ *queue.IngressQueue) {},
			ctx:        context.Background,
			msg:        msg,
			expect_err: nil,
		},
		{
			name: "returns context error when queue is full",
			setup: func(q *queue.IngressQueue) {
				q.Enqueue(context.Background(), msg)
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			msg:        msg,
			expect_err: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := queue.NewIngressQueue(1)
			tt.setup(q)

			err := q.Enqueue(tt.ctx(), tt.msg)

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("Enqueue() error = %v, want %v", err, tt.expect_err)
			}

			if tt.expect_err == nil {
				got, _ := q.Next(context.Background())
				if got != tt.msg {
					t.Errorf("Next() after Enqueue = %v, want %v", got, tt.msg)
				}
			}
		})
	}
}
