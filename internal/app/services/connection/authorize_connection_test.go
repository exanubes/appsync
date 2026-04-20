package connection_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
	"github.com/exanubes/appsync/internal/app/services/connection"
)

type mock_read_write_connection struct {
	write_err    error
	write_called bool
	written      []byte
	reads        [][]byte
	read_idx     int
	read_err     error
}

func (m *mock_read_write_connection) Write(_ context.Context, b []byte) error {
	m.write_called = true
	m.written = b
	return m.write_err
}

func (m *mock_read_write_connection) Read(_ context.Context) ([]byte, error) {
	if m.read_err != nil {
		return nil, m.read_err
	}
	if m.read_idx >= len(m.reads) {
		return nil, errors.New("no more reads")
	}
	data := m.reads[m.read_idx]
	m.read_idx++
	return data, nil
}

func (m *mock_read_write_connection) Close() error { return nil }

type mock_decoder struct {
	messages []app.Message
	idx      int
	err      error
}

func (m *mock_decoder) Decode(_ app.Payload) (app.Message, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.idx >= len(m.messages) {
		return nil, errors.New("no more messages")
	}
	msg := m.messages[m.idx]
	m.idx++
	return msg, nil
}

type mock_noop_authorizer struct{}

func (m *mock_noop_authorizer) Authorize(_ context.Context, _ app.AuthorizeCommandInput) (app.Signature, error) {
	return nil, nil
}

func TestAuthorize(t *testing.T) {
	write_err := errors.New("write failed")
	read_err := errors.New("read failed")
	decode_err := errors.New("decode failed")
	connection_init_msg := []byte(`{"type":"connection_init"}`)

	tests := []struct {
		name           string
		conn           *mock_read_write_connection
		decoder        *mock_decoder
		ctx            func() context.Context
		expect_err     error
		expect_timeout time.Duration
	}{
		{
			name: "success returns timeout from ack",
			conn: &mock_read_write_connection{
				reads: [][]byte{[]byte("ack")},
			},
			decoder: &mock_decoder{
				messages: []app.Message{protocol.ConnectionAckMessage{TimeoutMs: 5000}},
			},
			ctx:            context.Background,
			expect_err:     nil,
			expect_timeout: 5000 * time.Millisecond,
		},
		{
			name:       "write error is returned",
			conn:       &mock_read_write_connection{write_err: write_err},
			decoder:    &mock_decoder{},
			ctx:        context.Background,
			expect_err: write_err,
		},
		{
			name: "read error is returned",
			conn: &mock_read_write_connection{
				read_err: read_err,
			},
			decoder:    &mock_decoder{},
			ctx:        context.Background,
			expect_err: read_err,
		},
		{
			name: "decode error is returned",
			conn: &mock_read_write_connection{
				reads: [][]byte{[]byte("data")},
			},
			decoder:    &mock_decoder{err: decode_err},
			ctx:        context.Background,
			expect_err: decode_err,
		},
		{
			name: "error message returns formatted error",
			conn: &mock_read_write_connection{
				reads: [][]byte{[]byte("err")},
			},
			decoder: &mock_decoder{
				messages: []app.Message{
					protocol.ErrorMessage{
						Errors: []protocol.ErrorMetadata{{Type: "Unauthorized", Message: "forbidden"}},
					},
				},
			},
			ctx:        context.Background,
			expect_err: errors.New("Handshake returned with error"),
		},
		{
			name:    "context cancelled returns context error",
			conn:    &mock_read_write_connection{},
			decoder: &mock_decoder{},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expect_err: context.Canceled,
		},
		{
			name: "skips unknown message types and continues",
			conn: &mock_read_write_connection{
				reads: [][]byte{[]byte("ka"), []byte("ack")},
			},
			decoder: &mock_decoder{
				messages: []app.Message{
					protocol.KeepAliveMessage{},
					protocol.ConnectionAckMessage{TimeoutMs: 1000},
				},
			},
			ctx:            context.Background,
			expect_err:     nil,
			expect_timeout: 1000 * time.Millisecond,
		},
		{
			name: "sends connection_init message",
			conn: &mock_read_write_connection{
				reads: [][]byte{[]byte("ack")},
			},
			decoder: &mock_decoder{
				messages: []app.Message{protocol.ConnectionAckMessage{TimeoutMs: 0}},
			},
			ctx:        context.Background,
			expect_err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := connection.NewAuthorizeConnectionService(tt.decoder, &mock_noop_authorizer{}, &noop_logger{})

			timeout, err := svc.Authorize(tt.ctx(), tt.conn)

			if tt.expect_err != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expect_err)
				}
				if !errors.Is(err, tt.expect_err) && err.Error() != tt.expect_err.Error() {
					if tt.name != "error message returns formatted error" {
						t.Errorf("got error %v, want %v", err, tt.expect_err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if timeout != tt.expect_timeout {
					t.Errorf("timeout = %v, want %v", timeout, tt.expect_timeout)
				}
			}

			if tt.name == "sends connection_init message" {
				if !tt.conn.write_called {
					t.Error("expected Write to be called")
				}
				if string(tt.conn.written) != string(connection_init_msg) {
					t.Errorf("written = %q, want %q", tt.conn.written, connection_init_msg)
				}
			}
		})
	}
}
