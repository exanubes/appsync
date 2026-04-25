package appsync

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/subscription"
	"github.com/exanubes/appsync/internal/app/usecases/unsubscribe"
)

type mock_unsubscribe struct {
	err        error
	called     bool
	last_input unsubscribe.UnsubscribeChannelCommandInput
}

func (m *mock_unsubscribe) Execute(_ context.Context, input unsubscribe.UnsubscribeChannelCommandInput) error {
	m.called = true
	m.last_input = input
	return m.err
}

func make_channel_subscription(id string, sub *subscription.Subscription, u unsubscribe.UnsubscribeChannel) *ChannelSubscription {
	return &ChannelSubscription{
		id:           id,
		subscription: sub,
		unsubscribe:  u,
	}
}

func make_active_sub() *subscription.Subscription {
	sub, _ := subscription.New("sub-id", "test-channel", 1)
	return sub
}

func make_loaded_sub(payload app.Payload) *subscription.Subscription {
	sub, _ := subscription.New("sub-id", "test-channel", 1)
	sub.Deliver(context.Background(), payload)
	return sub
}

func make_closed_sub() *subscription.Subscription {
	sub, _ := subscription.New("sub-id", "test-channel", 1)
	sub.Close()
	return sub
}

func TestChannelSubscription_Close(t *testing.T) {
	sentinel_err := errors.New("unsubscribe failed")

	tests := []struct {
		name       string
		mock       *mock_unsubscribe
		expect_err error
	}{
		{
			name:       "returns nil on successful unsubscribe",
			mock:       &mock_unsubscribe{err: nil},
			expect_err: nil,
		},
		{
			name:       "propagates error from Execute",
			mock:       &mock_unsubscribe{err: sentinel_err},
			expect_err: sentinel_err,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := make_channel_subscription("id", make_active_sub(), tt.mock)

			err := cs.Close(context.Background())

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("Close() error = %v, want %v", err, tt.expect_err)
			}
			if !tt.mock.called {
				t.Error("Execute was not called")
			}
		})
	}
}

func TestChannelSubscription_Close_ForwardsInput(t *testing.T) {
	const sub_id = "my-sub-id"
	mock := &mock_unsubscribe{}
	cs := make_channel_subscription(sub_id, make_active_sub(), mock)

	cs.Close(context.Background())

	if mock.last_input.SubscriptionId != sub_id {
		t.Errorf("SubscriptionId = %q, want %q", mock.last_input.SubscriptionId, sub_id)
	}
	if mock.last_input.Frame == nil {
		t.Error("Frame must be non-nil")
	}
}

func TestChannelSubscription_Next(t *testing.T) {
	payload := app.Payload(`{"hello":"world"}`)

	tests := []struct {
		name        string
		sub         func() *subscription.Subscription
		ctx         func(*testing.T) context.Context
		expect_data []byte
		expect_err  error
	}{
		{
			name:        "returns NextMessageOutput with payload on success",
			sub:         func() *subscription.Subscription { return make_loaded_sub(payload) },
			ctx:         func(_ *testing.T) context.Context { return context.Background() },
			expect_data: payload,
			expect_err:  nil,
		},
		{
			name: "returns nil and context.Canceled when context is already cancelled",
			sub:  func() *subscription.Subscription { return make_active_sub() },
			ctx: func(_ *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expect_data: nil,
			expect_err:  context.Canceled,
		},
		{
			name: "returns nil and context.DeadlineExceeded when deadline expires",
			sub:  func() *subscription.Subscription { return make_active_sub() },
			ctx: func(t *testing.T) context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			expect_data: nil,
			expect_err:  context.DeadlineExceeded,
		},
		{
			name:        "returns nil and ErrSubscriptionClosed when subscription is closed",
			sub:         func() *subscription.Subscription { return make_closed_sub() },
			ctx:         func(_ *testing.T) context.Context { return context.Background() },
			expect_data: nil,
			expect_err:  app.ErrSubscriptionClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := make_channel_subscription("id", tt.sub(), &mock_unsubscribe{})

			got, err := cs.Next(tt.ctx(t))

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("Next() error = %v, want %v", err, tt.expect_err)
			}
			if tt.expect_data != nil {
				if got == nil {
					t.Fatal("Next() returned nil output, want non-nil")
				}
				if !bytes.Equal(got.Data, tt.expect_data) {
					t.Errorf("Next().Data = %v, want %v", got.Data, tt.expect_data)
				}
			} else if got != nil {
				t.Errorf("Next() = %v, want nil", got)
			}
		})
	}
}

func TestChannelSubscription_Next_UnblocksOnClose(t *testing.T) {
	sub, _ := subscription.New("sub-id", "test-channel", 0)
	cs := make_channel_subscription("id", sub, &mock_unsubscribe{})
	result := make(chan error, 1)

	go func() {
		_, err := cs.Next(context.Background())
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
		t.Error("Next() did not unblock after subscription.Close()")
	}
}

func TestChannelSubscription_DecodeNext(t *testing.T) {
	type message struct {
		Key string `json:"key"`
	}

	tests := []struct {
		name           string
		sub            func() *subscription.Subscription
		ctx            func(*testing.T) context.Context
		target         func() any
		check          func(*testing.T, any)
		expect_err     error
		expect_any_err bool
	}{
		{
			name:   "decodes valid JSON object into struct",
			sub:    func() *subscription.Subscription { return make_loaded_sub(app.Payload(`{"key":"value"}`)) },
			ctx:    func(_ *testing.T) context.Context { return context.Background() },
			target: func() any { return &message{} },
			check: func(t *testing.T, v any) {
				if got := v.(*message).Key; got != "value" {
					t.Errorf("Key = %q, want %q", got, "value")
				}
			},
		},
		{
			name:   "decodes valid JSON number into primitive",
			sub:    func() *subscription.Subscription { return make_loaded_sub(app.Payload("99")) },
			ctx:    func(_ *testing.T) context.Context { return context.Background() },
			target: func() any { n := 0; return &n },
			check: func(t *testing.T, v any) {
				if got := *v.(*int); got != 99 {
					t.Errorf("value = %d, want 99", got)
				}
			},
		},
		{
			name:           "returns error for invalid JSON and does not mutate target",
			sub:            func() *subscription.Subscription { return make_loaded_sub(app.Payload("not-valid-json")) },
			ctx:            func(_ *testing.T) context.Context { return context.Background() },
			target:         func() any { return &message{} },
			expect_any_err: true,
			check: func(t *testing.T, v any) {
				if got := v.(*message).Key; got != "" {
					t.Errorf("target was mutated: Key = %q, want empty", got)
				}
			},
		},
		{
			name: "returns context.Canceled when context is already cancelled",
			sub:  func() *subscription.Subscription { return make_active_sub() },
			ctx: func(_ *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			target:     func() any { return &message{} },
			expect_err: context.Canceled,
			check: func(t *testing.T, v any) {
				if got := v.(*message).Key; got != "" {
					t.Errorf("target was mutated: Key = %q, want empty", got)
				}
			},
		},
		{
			name:       "returns ErrSubscriptionClosed when subscription is closed",
			sub:        func() *subscription.Subscription { return make_closed_sub() },
			ctx:        func(_ *testing.T) context.Context { return context.Background() },
			target:     func() any { return &message{} },
			expect_err: app.ErrSubscriptionClosed,
			check:      func(_ *testing.T, _ any) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := make_channel_subscription("id", tt.sub(), &mock_unsubscribe{})
			target := tt.target()

			err := cs.DecodeNext(tt.ctx(t), target)

			if tt.expect_err != nil && !errors.Is(err, tt.expect_err) {
				t.Errorf("DecodeNext() error = %v, want %v", err, tt.expect_err)
			}
			if tt.expect_any_err && err == nil {
				t.Error("DecodeNext() expected error, got nil")
			}
			if !tt.expect_any_err && tt.expect_err == nil && err != nil {
				t.Errorf("DecodeNext() unexpected error = %v", err)
			}
			tt.check(t, target)
		})
	}
}
