package connection_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/services/connection"
)

type mock_authorizer struct {
	signature app.Signature
	err       error
	called    bool
	received  app.AuthorizeCommandInput
}

func (m *mock_authorizer) Authorize(_ context.Context, input app.AuthorizeCommandInput) (app.Signature, error) {
	m.called = true
	m.received = input
	return m.signature, m.err
}

type mock_serializer struct {
	result   string
	err      error
	called   bool
	received app.Signature
}

func (m *mock_serializer) Serialize(sig app.Signature) (string, error) {
	m.called = true
	m.received = sig
	return m.result, m.err
}

func TestGenerate(t *testing.T) {
	auth_err := errors.New("auth failed")
	serial_err := errors.New("serialize failed")
	signature := app.Signature{"Authorization": "sig-value"}

	tests := []struct {
		name              string
		authorizer        *mock_authorizer
		serializer        *mock_serializer
		expect_err        error
		expect_result     string
		expect_serialized bool
	}{
		{
			name:              "success",
			authorizer:        &mock_authorizer{signature: signature},
			serializer:        &mock_serializer{result: "encoded"},
			expect_err:        nil,
			expect_result:     "header-encoded",
			expect_serialized: true,
		},
		{
			name:              "authorizer error does not call serializer",
			authorizer:        &mock_authorizer{err: auth_err},
			serializer:        &mock_serializer{},
			expect_err:        auth_err,
			expect_serialized: false,
		},
		{
			name:              "serializer error is returned",
			authorizer:        &mock_authorizer{signature: signature},
			serializer:        &mock_serializer{err: serial_err},
			expect_err:        serial_err,
			expect_serialized: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := connection.NewGenerateSubprotocolService(tt.authorizer, tt.serializer)

			result, err := svc.Generate(context.Background())

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("got error %v, want %v", err, tt.expect_err)
			}

			if result != tt.expect_result {
				t.Errorf("got result %q, want %q", result, tt.expect_result)
			}

			if tt.authorizer.called {
				if tt.authorizer.received.Channel != "" || tt.authorizer.received.Payload != nil {
					t.Errorf("authorizer received %v, want empty AuthorizeCommandInput", tt.authorizer.received)
				}
			}

			if tt.serializer.called != tt.expect_serialized {
				t.Errorf("serializer.called = %v, want %v", tt.serializer.called, tt.expect_serialized)
			}

			if tt.expect_serialized {
				for k, v := range signature {
					if tt.serializer.received[k] != v {
						t.Errorf("serializer.received[%q] = %q, want %q", k, tt.serializer.received[k], v)
					}
				}
			}
		})
	}
}
