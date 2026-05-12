package io_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app"
	io_service "github.com/exanubes/appsync/internal/app/services/io"
)

type mock_outbox struct {
	payloads [][]byte
	idx      int
	err      error
}

func (m *mock_outbox) Next(_ context.Context) ([]byte, error) {
	if m.idx >= len(m.payloads) {
		return nil, m.err
	}
	p := m.payloads[m.idx]
	m.idx++
	return p, nil
}

type mock_connection struct {
	write_err    error
	write_called int
	written      [][]byte

	reads    [][]byte
	read_idx int
	read_err error
}

func (m *mock_connection) Write(_ context.Context, b []byte) error {
	if m.write_err != nil {
		return m.write_err
	}
	m.write_called++
	m.written = append(m.written, b)
	return nil
}

func (m *mock_connection) Read(_ context.Context) ([]byte, error) {
	if m.read_idx >= len(m.reads) {
		return nil, m.read_err
	}
	data := m.reads[m.read_idx]
	m.read_idx++
	return data, nil
}

func (m *mock_connection) Close(_ context.Context) error { return nil }

type mock_decoder struct {
	messages []app.Message
	idx      int
	err      error
}

func (m *mock_decoder) Decode(_ app.Payload) (app.Message, error) {
	if m.idx >= len(m.messages) {
		return nil, m.err
	}
	msg := m.messages[m.idx]
	m.idx++
	return msg, nil
}

type mock_inbox struct {
	enqueue_err   error
	enqueue_calls int
	enqueued      []app.Message
}

func (m *mock_inbox) Enqueue(_ context.Context, msg app.Message) error {
	m.enqueue_calls++
	m.enqueued = append(m.enqueued, msg)
	return m.enqueue_err
}

func TestWrite(t *testing.T) {
	outbox_err := errors.New("outbox exhausted")
	write_err := errors.New("write failed")
	payload_a := []byte("payload-a")
	payload_b := []byte("payload-b")

	tests := []struct {
		name          string
		outbox        *mock_outbox
		conn          *mock_connection
		ctx           func() context.Context
		expect_err    error
		expect_writes int
		verify        func(*testing.T, *mock_connection)
	}{
		{
			name:          "outbox error on first call",
			outbox:        &mock_outbox{err: outbox_err},
			conn:          &mock_connection{},
			ctx:           context.Background,
			expect_err:    outbox_err,
			expect_writes: 0,
		},
		{
			name:          "write error is returned",
			outbox:        &mock_outbox{payloads: [][]byte{payload_a}, err: outbox_err},
			conn:          &mock_connection{write_err: write_err},
			ctx:           context.Background,
			expect_err:    write_err,
			expect_writes: 0,
		},
		{
			name:          "payloads forwarded to connection in order",
			outbox:        &mock_outbox{payloads: [][]byte{payload_a, payload_b}, err: outbox_err},
			conn:          &mock_connection{},
			ctx:           context.Background,
			expect_err:    outbox_err,
			expect_writes: 2,
			verify: func(t *testing.T, conn *mock_connection) {
				if string(conn.written[0]) != string(payload_a) {
					t.Errorf("written[0] = %q, want %q", conn.written[0], payload_a)
				}
				if string(conn.written[1]) != string(payload_b) {
					t.Errorf("written[1] = %q, want %q", conn.written[1], payload_b)
				}
			},
		},
		{
			name:          "context cancellation propagates",
			outbox:        &mock_outbox{err: context.Canceled},
			conn:          &mock_connection{},
			ctx:           func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx },
			expect_err:    context.Canceled,
			expect_writes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := io_service.New(nil, tt.outbox, tt.conn, nil)
			err := svc.Write(tt.ctx())

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("got error %v, want %v", err, tt.expect_err)
			}
			if tt.conn.write_called != tt.expect_writes {
				t.Errorf("write_called = %d, want %d", tt.conn.write_called, tt.expect_writes)
			}
			if tt.verify != nil {
				tt.verify(t, tt.conn)
			}
		})
	}
}

func TestRead(t *testing.T) {
	read_err := errors.New("read failed")
	decode_err := errors.New("decode failed")
	enqueue_err := errors.New("enqueue failed")
	raw_a := []byte("raw-a")
	raw_b := []byte("raw-b")
	msg_a := "decoded-a"
	msg_b := "decoded-b"

	tests := []struct {
		name            string
		conn            *mock_connection
		decoder         *mock_decoder
		inbox           *mock_inbox
		ctx             func() context.Context
		expect_err      error
		expect_enqueued int
		verify          func(*testing.T, *mock_inbox)
	}{
		{
			name:            "read error returned immediately",
			conn:            &mock_connection{read_err: read_err},
			decoder:         &mock_decoder{err: decode_err},
			inbox:           &mock_inbox{},
			ctx:             context.Background,
			expect_err:      read_err,
			expect_enqueued: 0,
		},
		{
			name:            "decode error returned",
			conn:            &mock_connection{reads: [][]byte{raw_a}, read_err: read_err},
			decoder:         &mock_decoder{err: decode_err},
			inbox:           &mock_inbox{},
			ctx:             context.Background,
			expect_err:      decode_err,
			expect_enqueued: 0,
		},
		{
			name:            "enqueue error returned",
			conn:            &mock_connection{reads: [][]byte{raw_a}, read_err: read_err},
			decoder:         &mock_decoder{messages: []app.Message{msg_a}},
			inbox:           &mock_inbox{enqueue_err: enqueue_err},
			ctx:             context.Background,
			expect_err:      enqueue_err,
			expect_enqueued: 1,
		},
		{
			name:            "messages forwarded to inbox in order",
			conn:            &mock_connection{reads: [][]byte{raw_a, raw_b}, read_err: read_err},
			decoder:         &mock_decoder{messages: []app.Message{msg_a, msg_b}},
			inbox:           &mock_inbox{},
			ctx:             context.Background,
			expect_err:      read_err,
			expect_enqueued: 2,
			verify: func(t *testing.T, inbox *mock_inbox) {
				if inbox.enqueued[0] != msg_a {
					t.Errorf("enqueued[0] = %v, want %v", inbox.enqueued[0], msg_a)
				}
				if inbox.enqueued[1] != msg_b {
					t.Errorf("enqueued[1] = %v, want %v", inbox.enqueued[1], msg_b)
				}
			},
		},
		{
			name:            "context cancellation propagates",
			conn:            &mock_connection{read_err: context.Canceled},
			decoder:         &mock_decoder{err: decode_err},
			inbox:           &mock_inbox{},
			ctx:             func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx },
			expect_err:      context.Canceled,
			expect_enqueued: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := io_service.New(tt.inbox, nil, tt.conn, tt.decoder)
			err := svc.Read(tt.ctx())

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("got error %v, want %v", err, tt.expect_err)
			}
			if tt.inbox.enqueue_calls != tt.expect_enqueued {
				t.Errorf("enqueue_calls = %d, want %d", tt.inbox.enqueue_calls, tt.expect_enqueued)
			}
			if tt.verify != nil {
				tt.verify(t, tt.inbox)
			}
		})
	}
}
