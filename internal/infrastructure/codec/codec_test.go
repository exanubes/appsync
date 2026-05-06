package codec_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
	"github.com/exanubes/appsync/internal/infrastructure/codec"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name       string
		payload    app.Payload
		expect_msg app.Message
		expect_err bool
	}{
		{
			name:       "data message",
			payload:    []byte(`{"type":"data","id":"sub-1","event":"payload"}`),
			expect_msg: protocol.DataMessage{SubId: "sub-1", Payload: app.Payload("payload")},
		},
		{
			name:       "connection_ack message",
			payload:    []byte(`{"type":"connection_ack","connectionTimeoutMs":300000}`),
			expect_msg: protocol.ConnectionAckMessage{TimeoutMs: 300000},
		},
		{
			name:       "keep_alive message",
			payload:    []byte(`{"type":"ka"}`),
			expect_msg: protocol.KeepAliveMessage{},
		},
		{
			name:       "subscribe_success message",
			payload:    []byte(`{"type":"subscribe_success","id":"sub-1"}`),
			expect_msg: protocol.SuccessMessage{ID: "sub-1"},
		},
		{
			name:       "publish_success message",
			payload:    []byte(`{"type":"publish_success","id":"sub-1"}`),
			expect_msg: protocol.SuccessMessage{ID: "sub-1"},
		},
		{
			name:       "unsubscribe_success message",
			payload:    []byte(`{"type":"unsubscribe_success","id":"sub-1"}`),
			expect_msg: protocol.SuccessMessage{ID: "sub-1"},
		},
		{
			name:    "connection_error message",
			payload: []byte(`{"type":"connection_error","id":"sub-1","errors":[{"errorType":"Unauthorized","message":"not allowed"}]}`),
			expect_msg: protocol.ErrorMessage{
				ID:     "sub-1",
				Errors: []protocol.ErrorMetadata{{Type: "Unauthorized", Message: "not allowed"}},
			},
		},
		{
			name:    "error message",
			payload: []byte(`{"type":"error","id":"sub-1","errors":[{"errorType":"Unauthorized","message":"not allowed"}]}`),
			expect_msg: protocol.ErrorMessage{
				ID:     "sub-1",
				Errors: []protocol.ErrorMetadata{{Type: "Unauthorized", Message: "not allowed"}},
			},
		},
		{
			name:    "publish_error message",
			payload: []byte(`{"type":"publish_error","id":"sub-1","errors":[{"errorType":"Unauthorized","message":"not allowed"}]}`),
			expect_msg: protocol.ErrorMessage{
				ID:     "sub-1",
				Errors: []protocol.ErrorMetadata{{Type: "Unauthorized", Message: "not allowed"}},
			},
		},
		{
			name:    "subscribe_error message",
			payload: []byte(`{"type":"subscribe_error","id":"sub-1","errors":[{"errorType":"Unauthorized","message":"not allowed"}]}`),
			expect_msg: protocol.ErrorMessage{
				ID:     "sub-1",
				Errors: []protocol.ErrorMetadata{{Type: "Unauthorized", Message: "not allowed"}},
			},
		},
		{
			name:    "unsubscribe_error message",
			payload: []byte(`{"type":"unsubscribe_error","id":"sub-1","errors":[{"errorType":"Unauthorized","message":"not allowed"}]}`),
			expect_msg: protocol.ErrorMessage{
				ID:     "sub-1",
				Errors: []protocol.ErrorMetadata{{Type: "Unauthorized", Message: "not allowed"}},
			},
		},
		{
			name:       "unknown type returns envelope",
			payload:    []byte(`{"type":"unknown"}`),
			expect_msg: codec.Envelope{Type: "unknown"},
		},
		{
			name:       "invalid JSON returns error",
			payload:    []byte(`}{bad`),
			expect_err: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := codec.New()
			got, err := c.Decode(tt.payload)

			if tt.expect_err && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expect_err && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expect_msg != nil && !reflect.DeepEqual(got, tt.expect_msg) {
				t.Errorf("got %#v, want %#v", got, tt.expect_msg)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		name           string
		input          app.Message
		expect_payload app.Payload
		expect_err     bool
	}{
		{
			name:           "encodes struct to JSON",
			input:          protocol.ConnectionAckMessage{TimeoutMs: 5000},
			expect_payload: []byte(`{"TimeoutMs":5000}`),
		},
		{
			name:           "encodes empty struct to JSON",
			input:          protocol.KeepAliveMessage{},
			expect_payload: []byte(`{}`),
		},
		{
			name:       "returns error for unencodable value",
			input:      make(chan int),
			expect_err: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := codec.New()
			got, err := c.Encode(tt.input)

			if tt.expect_err && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expect_err && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expect_payload != nil && !bytes.Equal(got, tt.expect_payload) {
				t.Errorf("got %s, want %s", got, tt.expect_payload)
			}
		})
	}
}
