package queue_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app/queue"
)

type mock_connection_state struct{}

func (m *mock_connection_state) Done() <-chan struct{} { return make(chan struct{}) }

func TestEgressQueue_Next(t *testing.T) {
	payload := []byte("test-payload")

	tests := []struct {
		name           string
		setup          func(*queue.EgressQueue)
		ctx            func() context.Context
		expect_payload []byte
		expect_err     error
	}{
		{
			name: "returns payload when available",
			setup: func(q *queue.EgressQueue) {
				q.Enqueue(context.Background(), payload)
			},
			ctx:            context.Background,
			expect_payload: payload,
			expect_err:     nil,
		},
		{
			name:  "returns context error when context is cancelled",
			setup: func(_ *queue.EgressQueue) {},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expect_payload: nil,
			expect_err:     context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := queue.NewEgressQueue(1, &mock_connection_state{})
			tt.setup(q)

			got, err := q.Next(tt.ctx())

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("Next() error = %v, want %v", err, tt.expect_err)
			}
			if string(got) != string(tt.expect_payload) {
				t.Errorf("Next() = %v, want %v", got, tt.expect_payload)
			}
		})
	}
}

func TestEgressQueue_Enqueue(t *testing.T) {
	payload := []byte("test-payload")

	tests := []struct {
		name       string
		setup      func(*queue.EgressQueue)
		ctx        func() context.Context
		payload    []byte
		expect_err error
	}{
		{
			name:       "enqueues payload successfully",
			setup:      func(_ *queue.EgressQueue) {},
			ctx:        context.Background,
			payload:    payload,
			expect_err: nil,
		},
		{
			name: "returns context error when queue is full",
			setup: func(q *queue.EgressQueue) {
				q.Enqueue(context.Background(), payload)
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			payload:    payload,
			expect_err: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := queue.NewEgressQueue(1, &mock_connection_state{})
			tt.setup(q)

			err := q.Enqueue(tt.ctx(), tt.payload)

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("Enqueue() error = %v, want %v", err, tt.expect_err)
			}

			if tt.expect_err == nil {
				got, _ := q.Next(context.Background())
				if string(got) != string(tt.payload) {
					t.Errorf("Next() after Enqueue = %v, want %v", got, tt.payload)
				}
			}
		})
	}
}
