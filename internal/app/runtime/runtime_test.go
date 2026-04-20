package runtime_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/runtime"
)

type stub_inbox struct {
	messages  []app.Message
	idx       int
	empty_err error
}

func (s *stub_inbox) Next(ctx context.Context) (app.Message, error) {
	if s.idx >= len(s.messages) {
		if s.empty_err != nil {
			return nil, s.empty_err
		}
		<-ctx.Done()
		return nil, ctx.Err()
	}
	msg := s.messages[s.idx]
	s.idx++
	return msg, nil
}

type mock_router struct {
	err      error
	received []app.Message
}

func (m *mock_router) Handle(_ context.Context, msg app.Message) error {
	m.received = append(m.received, msg)
	return m.err
}

func TestRun(t *testing.T) {
	route_err := errors.New("route failed")
	msg_a := "message-a"
	msg_b := "message-b"

	tests := []struct {
		name             string
		inbox            *stub_inbox
		router           *mock_router
		ctx              func() context.Context
		expect_err       error
		expect_received  []app.Message
	}{
		{
			name: "routes all messages before returning inbox error",
			inbox: &stub_inbox{
				messages:  []app.Message{msg_a, msg_b},
				empty_err: context.Canceled,
			},
			router:          &mock_router{},
			ctx:             context.Background,
			expect_err:      context.Canceled,
			expect_received: []app.Message{msg_a, msg_b},
		},
		{
			name:            "returns router error",
			inbox:           &stub_inbox{messages: []app.Message{msg_a}},
			router:          &mock_router{err: route_err},
			ctx:             context.Background,
			expect_err:      route_err,
			expect_received: []app.Message{msg_a},
		},
		{
			name:  "returns context error when context is cancelled",
			inbox: &stub_inbox{},
			router: &mock_router{},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expect_err:      context.Canceled,
			expect_received: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := runtime.New(tt.router)
			err := r.Run(tt.ctx(), tt.inbox)

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("got error %v, want %v", err, tt.expect_err)
			}

			if len(tt.router.received) != len(tt.expect_received) {
				t.Fatalf("router received %d messages, want %d", len(tt.router.received), len(tt.expect_received))
			}

			for i, msg := range tt.expect_received {
				if tt.router.received[i] != msg {
					t.Errorf("router.received[%d] = %v, want %v", i, tt.router.received[i], msg)
				}
			}
		})
	}
}
