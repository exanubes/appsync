package router_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app/pending"
	"github.com/exanubes/appsync/internal/app/protocol"
	"github.com/exanubes/appsync/internal/app/router"
)

type mock_receive_data struct {
	called   bool
	received protocol.DataMessage
	err      error
}

func (m *mock_receive_data) Execute(_ context.Context, msg protocol.DataMessage) error {
	m.called = true
	m.received = msg
	return m.err
}

func TestHandle(t *testing.T) {
	const id = "sub-1"

	tests := []struct {
		name       string
		msg        any
		setup      func(*pending.Registry)
		receive    *mock_receive_data
		expect_err bool
		verify     func(*testing.T, *pending.Registry, *mock_receive_data)
	}{
		{
			name:    "error message without id returns error",
			msg:     protocol.ErrorMessage{Errors: []protocol.ErrorMetadata{{Type: "error", Message: "something went wrong"}}},
			receive: &mock_receive_data{},
			verify: func(t *testing.T, _ *pending.Registry, m *mock_receive_data) {
				if m.called {
					t.Error("receive use case should not be called")
				}
			},
			expect_err: true,
		},
		{
			name: "error message with id fulfills pending with error",
			msg:  protocol.ErrorMessage{ID: id, Errors: []protocol.ErrorMetadata{{Type: "error", Message: "something went wrong"}}},
			setup: func(r *pending.Registry) {
				r.Register(id)
			},
			receive: &mock_receive_data{},
			verify: func(t *testing.T, r *pending.Registry, _ *mock_receive_data) {
				consumed := r.Consume(context.Background(), id)
				if consumed == nil {
					t.Error("expected non-nil error from pending registry, got nil")
				}
			},
		},
		{
			name: "success message fulfills pending with nil",
			msg:  protocol.SuccessMessage{ID: id},
			setup: func(r *pending.Registry) {
				r.Register(id)
			},
			receive: &mock_receive_data{},
			verify: func(t *testing.T, r *pending.Registry, _ *mock_receive_data) {
				consumed := r.Consume(context.Background(), id)
				if consumed != nil {
					t.Errorf("expected nil error from pending registry, got %v", consumed)
				}
			},
		},
		{
			name:    "data message delegates to receive use case",
			msg:     protocol.DataMessage{SubId: id, Payload: []byte("data")},
			receive: &mock_receive_data{},
			verify: func(t *testing.T, _ *pending.Registry, m *mock_receive_data) {
				if !m.called {
					t.Error("receive use case was not called")
				}
				if m.received.SubId != id {
					t.Errorf("got SubId %q, want %q", m.received.SubId, id)
				}
			},
		},
		{
			name:    "data message propagates receive use case error",
			msg:     protocol.DataMessage{SubId: id, Payload: []byte("data")},
			receive: &mock_receive_data{err: errors.New("receive failed")},
			verify: func(t *testing.T, _ *pending.Registry, m *mock_receive_data) {
				if !m.called {
					t.Error("receive use case was not called")
				}
			},
			expect_err: true,
		},
		{
			name:    "unknown message type returns nil",
			msg:     struct{ Foo string }{Foo: "bar"},
			receive: &mock_receive_data{},
			verify: func(t *testing.T, _ *pending.Registry, m *mock_receive_data) {
				if m.called {
					t.Error("receive use case should not be called for unknown message type")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := pending.NewRegistry()
			if tt.setup != nil {
				tt.setup(registry)
			}

			handler := router.New(registry, tt.receive)
			err := handler.Handle(context.Background(), tt.msg)

			if tt.expect_err && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expect_err && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.verify != nil {
				tt.verify(t, registry, tt.receive)
			}
		})
	}
}
