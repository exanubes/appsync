package subscription

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/exanubes/appsync/internal/app"
)

const test_enqueue_timeout = 5 * time.Millisecond

func make_full_inbox(payload app.Payload) *inbox {
	i := new_inbox(1)
	i.timeout = test_enqueue_timeout
	i.Enqueue(context.Background(), payload)
	return i
}

func TestInbox_Enqueue(t *testing.T) {
	payload := app.Payload("test-payload")

	tests := []struct {
		name       string
		inbox      func() *inbox
		ctx        func(*testing.T) context.Context
		expect_err error
	}{
		{
			name:       "enqueues payload successfully",
			inbox:      func() *inbox { return new_inbox(1) },
			ctx:        func(_ *testing.T) context.Context { return context.Background() },
			expect_err: nil,
		},
		{
			name:       "returns ErrSubscriptionInboxFull when inbox is full",
			inbox:      func() *inbox { return make_full_inbox(payload) },
			ctx:        func(_ *testing.T) context.Context { return context.Background() },
			expect_err: app.ErrSubscriptionInboxFull,
		},
		{
			name:  "returns context.Canceled when context is already cancelled",
			inbox: func() *inbox { return new_inbox(1) },
			ctx: func(_ *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expect_err: context.Canceled,
		},
		{
			name:  "returns context.DeadlineExceeded when deadline expires before inbox timeout",
			inbox: func() *inbox { return make_full_inbox(payload) },
			ctx: func(t *testing.T) context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			expect_err: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.inbox().Enqueue(tt.ctx(t), payload)

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("Enqueue() error = %v, want %v", err, tt.expect_err)
			}
		})
	}
}

func TestInbox_Next(t *testing.T) {
	payload := app.Payload("test-payload")

	tests := []struct {
		name           string
		setup          func(*inbox)
		ctx            func(*testing.T) context.Context
		expect_payload app.Payload
		expect_err     error
	}{
		{
			name: "returns payload when available",
			setup: func(i *inbox) {
				i.Enqueue(context.Background(), payload)
			},
			ctx:            func(_ *testing.T) context.Context { return context.Background() },
			expect_payload: payload,
			expect_err:     nil,
		},
		{
			name:  "returns context.Canceled when context is already cancelled",
			setup: func(_ *inbox) {},
			ctx: func(_ *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expect_payload: nil,
			expect_err:     context.Canceled,
		},
		{
			name:  "returns context.DeadlineExceeded when context deadline expires",
			setup: func(_ *inbox) {},
			ctx: func(t *testing.T) context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			expect_payload: nil,
			expect_err:     context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := new_inbox(1)
			tt.setup(i)

			got, err := i.Next(tt.ctx(t))

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("Next() error = %v, want %v", err, tt.expect_err)
			}
			if !bytes.Equal(got, tt.expect_payload) {
				t.Errorf("Next() = %v, want %v", got, tt.expect_payload)
			}
		})
	}
}

func TestInbox_Next_FIFOOrder(t *testing.T) {
	first := app.Payload("first")
	second := app.Payload("second")

	i := new_inbox(2)
	i.Enqueue(context.Background(), first)
	i.Enqueue(context.Background(), second)

	got1, err := i.Next(context.Background())
	if err != nil {
		t.Fatalf("first Next() error = %v", err)
	}
	got2, err := i.Next(context.Background())
	if err != nil {
		t.Fatalf("second Next() error = %v", err)
	}

	if !bytes.Equal(got1, first) {
		t.Errorf("first Next() = %v, want %v", got1, first)
	}
	if !bytes.Equal(got2, second) {
		t.Errorf("second Next() = %v, want %v", got2, second)
	}
}
