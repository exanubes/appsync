package connection_test

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"testing"
	"time"

	"github.com/exanubes/appsync/internal/app/services/connection"
)

type mock_subprotocol_generator struct {
	result string
	err    error
	called bool
}

func (m *mock_subprotocol_generator) Generate(_ context.Context) (string, error) {
	m.called = true
	return m.result, m.err
}

type mock_dialer struct {
	conn     connection.Connection
	err      error
	called   bool
	received connection.DialOptions
}

func (m *mock_dialer) Dial(_ context.Context, opts connection.DialOptions) (connection.Connection, error) {
	m.called = true
	m.received = opts
	return m.conn, m.err
}

type mock_connection_authorizer struct {
	timeout  time.Duration
	err      error
	called   bool
	received connection.Connection
}

func (m *mock_connection_authorizer) Authorize(_ context.Context, conn connection.Connection) (time.Duration, error) {
	m.called = true
	m.received = conn
	return m.timeout, m.err
}

type mock_connection struct{}

func (m *mock_connection) Read(_ context.Context) ([]byte, error)  { return nil, nil }
func (m *mock_connection) Write(_ context.Context, _ []byte) error { return nil }
func (m *mock_connection) Close(_ context.Context) error           { return nil }

func TestConnect(t *testing.T) {
	gen_err := errors.New("generate failed")
	dial_err := errors.New("dial failed")
	auth_err := errors.New("authorize failed")
	conn := &mock_connection{}
	timeout := 5 * time.Second
	endpoint, _ := url.Parse("wss://example.com/graphql")

	tests := []struct {
		name              string
		generator         *mock_subprotocol_generator
		dialer            *mock_dialer
		authorizer        *mock_connection_authorizer
		input             connection.CreateConnectionInput
		expect_err        error
		expect_dialed     bool
		expect_authorized bool
		verify            func(*testing.T, *mock_dialer, *connection.CreateConnectionOutput)
	}{
		{
			name:              "success",
			generator:         &mock_subprotocol_generator{result: "header-xyz"},
			dialer:            &mock_dialer{conn: conn},
			authorizer:        &mock_connection_authorizer{timeout: timeout},
			input:             connection.CreateConnectionInput{Url: endpoint},
			expect_dialed:     true,
			expect_authorized: true,
			verify: func(t *testing.T, dialer *mock_dialer, output *connection.CreateConnectionOutput) {
				if dialer.received.Url != endpoint {
					t.Errorf("dialer.received.Url = %v, want %v", dialer.received.Url, endpoint)
				}
				got := dialer.received.Subprotocols
				if len(got) == 0 || got[len(got)-1] != "header-xyz" {
					t.Errorf("dialer subprotocols last element = %v, want %q", got, "header-xyz")
				}
				if output == nil {
					t.Fatal("expected non-nil output")
				}
				if output.Connection != conn {
					t.Errorf("output.Connection = %v, want %v", output.Connection, conn)
				}
				if output.Timeout != timeout {
					t.Errorf("output.Timeout = %v, want %v", output.Timeout, timeout)
				}
			},
		},
		{
			name:       "subprotocol error does not call dialer",
			generator:  &mock_subprotocol_generator{err: gen_err},
			dialer:     &mock_dialer{},
			authorizer: &mock_connection_authorizer{},
			input:      connection.CreateConnectionInput{Url: endpoint},
			expect_err: gen_err,
		},
		{
			name:          "dialer error does not call authorizer",
			generator:     &mock_subprotocol_generator{result: "header-xyz"},
			dialer:        &mock_dialer{err: dial_err},
			authorizer:    &mock_connection_authorizer{},
			input:         connection.CreateConnectionInput{Url: endpoint},
			expect_err:    dial_err,
			expect_dialed: true,
			verify: func(t *testing.T, dialer *mock_dialer, _ *connection.CreateConnectionOutput) {
				if dialer.received.Url != endpoint {
					t.Errorf("dialer.received.Url = %v, want %v", dialer.received.Url, endpoint)
				}
				got := dialer.received.Subprotocols
				if len(got) == 0 || got[len(got)-1] != "header-xyz" {
					t.Errorf("dialer subprotocols last element = %v, want %q", got, "header-xyz")
				}
			},
		},
		{
			name:              "authorizer error is returned",
			generator:         &mock_subprotocol_generator{result: "header-xyz"},
			dialer:            &mock_dialer{conn: conn},
			authorizer:        &mock_connection_authorizer{err: auth_err},
			input:             connection.CreateConnectionInput{Url: endpoint},
			expect_err:        auth_err,
			expect_dialed:     true,
			expect_authorized: true,
			verify: func(t *testing.T, dialer *mock_dialer, _ *connection.CreateConnectionOutput) {
				if dialer.received.Url != endpoint {
					t.Errorf("dialer.received.Url = %v, want %v", dialer.received.Url, endpoint)
				}
				got := dialer.received.Subprotocols
				if len(got) == 0 || got[len(got)-1] != "header-xyz" {
					t.Errorf("dialer subprotocols last element = %v, want %q", got, "header-xyz")
				}
			},
		},
		{
			name:      "generated subprotocol appended to input subprotocols",
			generator: &mock_subprotocol_generator{result: "header-xyz"},
			dialer:    &mock_dialer{conn: conn},
			authorizer: &mock_connection_authorizer{timeout: timeout},
			input: connection.CreateConnectionInput{
				Url:          endpoint,
				Subprotocols: []string{"header-abc"},
			},
			expect_dialed:     true,
			expect_authorized: true,
			verify: func(t *testing.T, dialer *mock_dialer, output *connection.CreateConnectionOutput) {
				got := dialer.received.Subprotocols
				if len(got) == 0 || got[len(got)-1] != "header-xyz" {
					t.Errorf("dialer subprotocols last element = %v, want %q", got, "header-xyz")
				}
				if !slices.Contains(got, "header-abc") {
					t.Errorf("input subprotocol %q missing from dialer received %v", "header-abc", got)
				}
				if output == nil {
					t.Fatal("expected non-nil output")
				}
			},
		},
		{
			name:              "input URL forwarded to dialer",
			generator:         &mock_subprotocol_generator{result: "header-xyz"},
			dialer:            &mock_dialer{conn: conn},
			authorizer:        &mock_connection_authorizer{timeout: timeout},
			input:             connection.CreateConnectionInput{Url: endpoint},
			expect_dialed:     true,
			expect_authorized: true,
			verify: func(t *testing.T, dialer *mock_dialer, _ *connection.CreateConnectionOutput) {
				if dialer.received.Url != endpoint {
					t.Errorf("dialer.received.Url = %v, want %v", dialer.received.Url, endpoint)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := connection.NewConnectionService(tt.dialer, tt.authorizer, tt.generator, &noop_logger{})

			output, err := svc.Connect(context.Background(), tt.input)

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("got error %v, want %v", err, tt.expect_err)
			}
			if tt.dialer.called != tt.expect_dialed {
				t.Errorf("dialer.called = %v, want %v", tt.dialer.called, tt.expect_dialed)
			}
			if tt.authorizer.called != tt.expect_authorized {
				t.Errorf("authorizer.called = %v, want %v", tt.authorizer.called, tt.expect_authorized)
			}
			if tt.verify != nil {
				tt.verify(t, tt.dialer, output)
			}
		})
	}
}
