package pending_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app/pending"
)

func TestHas(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*pending.Registry)
		id         string
		expect_has bool
	}{
		{
			name:       "returns false when id not registered",
			setup:      func(_ *pending.Registry) {},
			id:         "test-id",
			expect_has: false,
		},
		{
			name: "returns true after Register",
			setup: func(r *pending.Registry) {
				r.Register("test-id")
			},
			id:         "test-id",
			expect_has: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := pending.NewRegistry()
			tt.setup(registry)
			has := registry.Has(tt.id)
			if has != tt.expect_has {
				t.Errorf("Has(%q) = %v, want %v", tt.id, has, tt.expect_has)
			}
		})
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "registered id is available via Has",
			id:   "test-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := pending.NewRegistry()
			registry.Register(tt.id)
			if !registry.Has(tt.id) {
				t.Errorf("Has(%q) = false after Register, want true", tt.id)
			}
		})
	}
}

func TestFulfill(t *testing.T) {
	fulfill_err := errors.New("fulfill error")

	tests := []struct {
		name       string
		setup      func(*pending.Registry)
		id         string
		err        error
		ctx        func() context.Context
		expect_err error
	}{
		{
			name:       "returns nil when id not registered",
			setup:      func(_ *pending.Registry) {},
			id:         "unknown-id",
			err:        nil,
			ctx:        context.Background,
			expect_err: nil,
		},
		{
			name: "sends nil error to channel",
			setup: func(r *pending.Registry) {
				r.Register("test-id")
			},
			id:         "test-id",
			err:        nil,
			ctx:        context.Background,
			expect_err: nil,
		},
		{
			name: "sends error to channel",
			setup: func(r *pending.Registry) {
				r.Register("test-id")
			},
			id:         "test-id",
			err:        fulfill_err,
			ctx:        context.Background,
			expect_err: nil,
		},
		{
			name: "context cancelled returns context error",
			setup: func(r *pending.Registry) {
				r.Register("test-id")
				r.Fulfill(context.Background(), "test-id", nil)
			},
			id:  "test-id",
			err: fulfill_err,
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expect_err: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := pending.NewRegistry()
			tt.setup(registry)
			err := registry.Fulfill(tt.ctx(), tt.id, tt.err)
			if !errors.Is(err, tt.expect_err) {
				t.Errorf("Fulfill() error = %v, want %v", err, tt.expect_err)
			}
		})
	}
}

func TestConsume(t *testing.T) {
	consume_err := errors.New("consume error")

	tests := []struct {
		name             string
		setup            func(*pending.Registry)
		id               string
		ctx              func() context.Context
		expect_err       error
		expect_has_after bool
	}{
		{
			name:             "returns nil when id not registered",
			setup:            func(_ *pending.Registry) {},
			id:               "unknown-id",
			ctx:              context.Background,
			expect_err:       nil,
			expect_has_after: false,
		},
		{
			name: "receives nil error from channel",
			setup: func(r *pending.Registry) {
				r.Register("test-id")
				r.Fulfill(context.Background(), "test-id", nil)
			},
			id:               "test-id",
			ctx:              context.Background,
			expect_err:       nil,
			expect_has_after: false,
		},
		{
			name: "receives error from channel",
			setup: func(r *pending.Registry) {
				r.Register("test-id")
				r.Fulfill(context.Background(), "test-id", consume_err)
			},
			id:               "test-id",
			ctx:              context.Background,
			expect_err:       consume_err,
			expect_has_after: false,
		},
		{
			name: "deletes entry after consuming",
			setup: func(r *pending.Registry) {
				r.Register("test-id")
				r.Fulfill(context.Background(), "test-id", nil)
			},
			id:               "test-id",
			ctx:              context.Background,
			expect_err:       nil,
			expect_has_after: false,
		},
		{
			name: "context cancelled returns context error",
			setup: func(r *pending.Registry) {
				r.Register("test-id")
			},
			id: "test-id",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expect_err:       context.Canceled,
			expect_has_after: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := pending.NewRegistry()
			tt.setup(registry)
			err := registry.Consume(tt.ctx(), tt.id)
			if !errors.Is(err, tt.expect_err) {
				t.Errorf("Consume() error = %v, want %v", err, tt.expect_err)
			}
			has := registry.Has(tt.id)
			if has != tt.expect_has_after {
				t.Errorf("Has(%q) after Consume = %v, want %v", tt.id, has, tt.expect_has_after)
			}
		})
	}
}
