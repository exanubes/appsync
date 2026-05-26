package request_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/services/request"
)

type mock_queue struct {
	err      error
	called   bool
	received []byte
}

func (m *mock_queue) Enqueue(_ context.Context, p []byte) error {
	m.called = true
	m.received = p
	return m.err
}

type mock_registry struct {
	has         bool
	consume_err error
	register_id string
}

func (m *mock_registry) Has(_ string) bool  { return m.has }
func (m *mock_registry) Register(id string) { m.register_id = id }
func (m *mock_registry) Remove(id string) {
	if m.register_id == id {
		m.register_id = ""
	}
}
func (m *mock_registry) Consume(_ context.Context, id string) error {
	return m.consume_err
}

type mock_frame struct {
	id         string
	payload    app.Payload
	encode_err error
}

func (m *mock_frame) ID() string                   { return m.id }
func (m *mock_frame) Encode() (app.Payload, error) { return m.payload, m.encode_err }

func TestSend(t *testing.T) {
	encode_err := errors.New("encode failed")
	enqueue_err := errors.New("enqueue failed")
	consume_err := errors.New("consume failed")
	payload := app.Payload("test-payload")
	frame_id := "test-id"

	tests := []struct {
		name              string
		registry          *mock_registry
		queue             *mock_queue
		frame             *mock_frame
		expect_err        error
		expect_enqueued   bool
		expect_registered bool
	}{
		{
			name:              "duplicate message returns error",
			registry:          &mock_registry{has: true},
			queue:             &mock_queue{},
			frame:             &mock_frame{id: frame_id, payload: payload},
			expect_err:        app.ErrDuplicateMessage,
			expect_enqueued:   false,
			expect_registered: false,
		},
		{
			name:              "encode error is returned",
			registry:          &mock_registry{},
			queue:             &mock_queue{},
			frame:             &mock_frame{id: frame_id, encode_err: encode_err},
			expect_err:        encode_err,
			expect_enqueued:   false,
			expect_registered: false,
		},
		{
			name:              "enqueue error is returned",
			registry:          &mock_registry{},
			queue:             &mock_queue{err: enqueue_err},
			frame:             &mock_frame{id: frame_id, payload: payload},
			expect_err:        enqueue_err,
			expect_enqueued:   true,
			expect_registered: false,
		},
		{
			name:              "success",
			registry:          &mock_registry{},
			queue:             &mock_queue{},
			frame:             &mock_frame{id: frame_id, payload: payload},
			expect_err:        nil,
			expect_enqueued:   true,
			expect_registered: true,
		},
		{
			name:              "consume error is returned",
			registry:          &mock_registry{consume_err: consume_err},
			queue:             &mock_queue{},
			frame:             &mock_frame{id: frame_id, payload: payload},
			expect_err:        consume_err,
			expect_enqueued:   true,
			expect_registered: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := request.NewSendRequestService(tt.queue, tt.registry)

			err := svc.Send(context.Background(), tt.frame)

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("got error %v, want %v", err, tt.expect_err)
			}

			if tt.queue.called != tt.expect_enqueued {
				t.Errorf("queue.called = %v, want %v", tt.queue.called, tt.expect_enqueued)
			}

			if tt.expect_enqueued && string(tt.queue.received) != string(tt.frame.payload) {
				t.Errorf("queue.received = %q, want %q", tt.queue.received, tt.frame.payload)
			}

			registered := tt.registry.register_id == frame_id
			if registered != tt.expect_registered {
				t.Errorf("registered = %v, want %v", registered, tt.expect_registered)
			}
		})
	}
}
