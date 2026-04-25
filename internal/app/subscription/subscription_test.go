package subscription

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/exanubes/appsync/internal/app"
)

func make_full_subscription(payload app.Payload) *Subscription {
	sub, _ := New("id", "ch", 1)
	sub.inbox.timeout = test_enqueue_timeout
	sub.Deliver(context.Background(), payload)
	return sub
}

func TestSubscription_New(t *testing.T) {
	payload := app.Payload("test-payload")

	t.Run("can deliver and receive a payload", func(t *testing.T) {
		sub, _ := New("test-id", "test-channel", 1)

		if err := sub.Deliver(context.Background(), payload); err != nil {
			t.Fatalf("Deliver() error = %v", err)
		}

		got, err := sub.Next(context.Background())
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if !bytes.Equal(got, payload) {
			t.Errorf("Next() = %v, want %v", got, payload)
		}
	})

	t.Run("zero buffer forces synchronous rendezvous", func(t *testing.T) {
		sub, _ := New("test-id", "test-channel", 0)
		received := make(chan app.Payload, 1)

		go func() {
			got, _ := sub.Next(context.Background())
			received <- got
		}()

		time.Sleep(10 * time.Millisecond)

		if err := sub.Deliver(context.Background(), payload); err != nil {
			t.Fatalf("Deliver() error = %v", err)
		}

		select {
		case got := <-received:
			if !bytes.Equal(got, payload) {
				t.Errorf("Next() = %v, want %v", got, payload)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Next() did not unblock after Deliver")
		}
	})
}

func TestSubscription_Deliver(t *testing.T) {
	payload := app.Payload("test-payload")

	tests := []struct {
		name       string
		sub        func() *Subscription
		ctx        func(*testing.T) context.Context
		expect_err error
	}{
		{
			name:       "delivers payload successfully",
			sub:        func() *Subscription { sub, _ := New("id", "ch", 1); return sub },
			ctx:        func(_ *testing.T) context.Context { return context.Background() },
			expect_err: nil,
		},
		{
			name:       "returns ErrSubscriptionInboxFull when inbox is full",
			sub:        func() *Subscription { return make_full_subscription(payload) },
			ctx:        func(_ *testing.T) context.Context { return context.Background() },
			expect_err: app.ErrSubscriptionInboxFull,
		},
		{
			name: "returns context.Canceled when context is already cancelled",
			sub:  func() *Subscription { return make_full_subscription(payload) },
			ctx: func(_ *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expect_err: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sub().Deliver(tt.ctx(t), payload)

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("Deliver() error = %v, want %v", err, tt.expect_err)
			}
		})
	}
}

func TestSubscription_Next(t *testing.T) {
	payload := app.Payload("test-payload")

	tests := []struct {
		name           string
		setup          func(*Subscription)
		ctx            func(*testing.T) context.Context
		expect_payload []byte
		expect_err     error
	}{
		{
			name: "returns payload when available",
			setup: func(sub *Subscription) {
				sub.Deliver(context.Background(), payload)
			},
			ctx:            func(_ *testing.T) context.Context { return context.Background() },
			expect_payload: payload,
			expect_err:     nil,
		},
		{
			name:  "returns context.Canceled when context is already cancelled",
			setup: func(_ *Subscription) {},
			ctx: func(_ *testing.T) context.Context {
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
			sub, _ := New("id", "ch", 1)
			tt.setup(sub)

			got, err := sub.Next(tt.ctx(t))

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("Next() error = %v, want %v", err, tt.expect_err)
			}
			if !bytes.Equal(got, tt.expect_payload) {
				t.Errorf("Next() = %v, want %v", got, tt.expect_payload)
			}
		})
	}
}

func TestSubscription_Next_ConcurrentDelivery(t *testing.T) {
	payload := app.Payload("test-payload")
	sub, _ := New("id", "ch", 0)
	received := make(chan app.Payload, 1)

	go func() {
		got, _ := sub.Next(context.Background())
		received <- got
	}()

	time.Sleep(10 * time.Millisecond)

	if err := sub.Deliver(context.Background(), payload); err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}

	select {
	case got := <-received:
		if !bytes.Equal(got, payload) {
			t.Errorf("Next() = %v, want %v", got, payload)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Next() did not unblock after concurrent Deliver")
	}
}

func TestSubscription_Close(t *testing.T) {
	t.Run("is idempotent — multiple calls do not panic", func(t *testing.T) {
		sub, _ := New("id", "ch", 1)
		sub.Close()
		sub.Close()
	})

	t.Run("Next() returns ErrSubscriptionClosed after close", func(t *testing.T) {
		sub, _ := New("id", "ch", 1)
		sub.Close()

		_, err := sub.Next(context.Background())
		if !errors.Is(err, app.ErrSubscriptionClosed) {
			t.Errorf("Next() error = %v, want %v", err, app.ErrSubscriptionClosed)
		}
	})

	t.Run("Deliver() returns ErrSubscriptionClosed after close", func(t *testing.T) {
		sub, _ := New("id", "ch", 0)
		sub.Close()

		err := sub.Deliver(context.Background(), app.Payload("test"))
		if !errors.Is(err, app.ErrSubscriptionClosed) {
			t.Errorf("Deliver() error = %v, want %v", err, app.ErrSubscriptionClosed)
		}
	})

	t.Run("unblocks pending Next() with ErrSubscriptionClosed", func(t *testing.T) {
		sub, _ := New("id", "ch", 0)
		result := make(chan error, 1)

		go func() {
			_, err := sub.Next(context.Background())
			result <- err
		}()

		time.Sleep(10 * time.Millisecond)
		sub.Close()

		select {
		case err := <-result:
			if !errors.Is(err, app.ErrSubscriptionClosed) {
				t.Errorf("Next() error = %v, want %v", err, app.ErrSubscriptionClosed)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Next() did not unblock after Close()")
		}
	})
}

func TestSubscription_Active(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*Subscription)
		expect bool
	}{
		{
			name:   "returns true for a new subscription",
			setup:  func(_ *Subscription) {},
			expect: true,
		},
		{
			name:   "returns false after Close()",
			setup:  func(sub *Subscription) { sub.Close() },
			expect: false,
		},
		{
			name: "remains false after redundant Close()",
			setup: func(sub *Subscription) {
				sub.Close()
				sub.Close()
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, _ := New("id", "ch", 1)
			tt.setup(sub)

			if got := sub.Active(); got != tt.expect {
				t.Errorf("Active() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestSubscription_Decode(t *testing.T) {
	type json_target struct {
		Key string `json:"key"`
	}

	tests := []struct {
		name       string
		setup      func(*Subscription)
		ctx        func(*testing.T) context.Context
		target     func() any
		check      func(*testing.T, any)
		expect_err bool
	}{
		{
			name: "decodes valid JSON payload into target struct",
			setup: func(sub *Subscription) {
				sub.Deliver(context.Background(), app.Payload(`{"key":"value"}`))
			},
			ctx:    func(_ *testing.T) context.Context { return context.Background() },
			target: func() any { return &json_target{} },
			check: func(t *testing.T, v any) {
				if got := v.(*json_target).Key; got != "value" {
					t.Errorf("Key = %q, want %q", got, "value")
				}
			},
			expect_err: false,
		},
		{
			name: "decodes JSON number into target",
			setup: func(sub *Subscription) {
				sub.Deliver(context.Background(), app.Payload("42"))
			},
			ctx:    func(_ *testing.T) context.Context { return context.Background() },
			target: func() any { v := 0; return &v },
			check: func(t *testing.T, v any) {
				if got := *v.(*int); got != 42 {
					t.Errorf("value = %d, want 42", got)
				}
			},
			expect_err: false,
		},
		{
			name: "returns error for invalid JSON",
			setup: func(sub *Subscription) {
				sub.Deliver(context.Background(), app.Payload("not-json"))
			},
			ctx:        func(_ *testing.T) context.Context { return context.Background() },
			target:     func() any { return &json_target{} },
			check:      func(_ *testing.T, _ any) {},
			expect_err: true,
		},
		{
			name:  "returns context.Canceled and does not mutate target when context is cancelled",
			setup: func(_ *Subscription) {},
			ctx: func(_ *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			target: func() any { return &json_target{} },
			check: func(t *testing.T, v any) {
				if got := v.(*json_target).Key; got != "" {
					t.Errorf("target was mutated, Key = %q, want empty", got)
				}
			},
			expect_err: true,
		},
		{
			name:  "returns context.DeadlineExceeded when deadline expires on empty inbox",
			setup: func(_ *Subscription) {},
			ctx: func(t *testing.T) context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			target:     func() any { return &json_target{} },
			check:      func(_ *testing.T, _ any) {},
			expect_err: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, _ := New("id", "ch", 1)
			tt.setup(sub)

			target := tt.target()
			err := sub.Decode(tt.ctx(t), target)

			if tt.expect_err && err == nil {
				t.Error("Decode() expected error, got nil")
			}
			if !tt.expect_err && err != nil {
				t.Errorf("Decode() unexpected error = %v", err)
			}

			tt.check(t, target)
		})
	}
}
