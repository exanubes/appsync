package shutdown_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app/usecases/shutdown"
)

type mock_registry struct {
	active []string
}

func (m *mock_registry) Active() []string { return m.active }

type mock_remover struct {
	err          error
	called       bool
	received_ids []string
}

func (m *mock_remover) Remove(_ context.Context, ids ...string) error {
	m.called = true
	m.received_ids = ids
	return m.err
}

type mock_closer struct {
	err    error
	called bool
}

func (m *mock_closer) Close(_ context.Context) error {
	m.called = true
	return m.err
}

func TestShutdownConnection(t *testing.T) {
	sentinel_unsub := errors.New("remove failed")
	sentinel_conn := errors.New("runtime close failed")
	sentinel_trans := errors.New("transport close failed")

	tests := []struct {
		name           string
		registry_ids   []string
		remover_err    error
		runtime_err    error
		transport_err  error
		expect_errors  []error
	}{
		{
			name:          "success",
			registry_ids:  []string{"a", "b"},
			expect_errors: nil,
		},
		{
			name:          "no active subscriptions",
			registry_ids:  []string{},
			expect_errors: nil,
		},
		{
			name:          "subscription removal error",
			registry_ids:  []string{"a"},
			remover_err:   sentinel_unsub,
			expect_errors: []error{sentinel_unsub},
		},
		{
			name:          "runtime close error",
			registry_ids:  []string{"a"},
			runtime_err:   sentinel_conn,
			expect_errors: []error{sentinel_conn},
		},
		{
			name:          "transport close error",
			registry_ids:  []string{"a"},
			transport_err: sentinel_trans,
			expect_errors: []error{sentinel_trans},
		},
		{
			name:          "all errors joined",
			registry_ids:  []string{"a"},
			remover_err:   sentinel_unsub,
			runtime_err:   sentinel_conn,
			transport_err: sentinel_trans,
			expect_errors: []error{sentinel_unsub, sentinel_conn, sentinel_trans},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &mock_registry{active: tt.registry_ids}
			remover := &mock_remover{err: tt.remover_err}
			runtime := &mock_closer{err: tt.runtime_err}
			transport := &mock_closer{err: tt.transport_err}
			usecase := shutdown.NewShutdownConnectionUsecase(registry, remover, runtime, transport)

			err := usecase.Execute(context.Background())

			if len(tt.expect_errors) == 0 {
				if err != nil {
					t.Errorf("got error %v, want nil", err)
				}
				return
			}
			for _, expected := range tt.expect_errors {
				if !errors.Is(err, expected) {
					t.Errorf("error %v not found in %v", expected, err)
				}
			}
		})
	}
}

func TestShutdownConnection_ForwardsActiveSubscriptions(t *testing.T) {
	ids := []string{"sub-1", "sub-2", "sub-3"}
	registry := &mock_registry{active: ids}
	remover := &mock_remover{}
	usecase := shutdown.NewShutdownConnectionUsecase(registry, remover, &mock_closer{}, &mock_closer{})

	usecase.Execute(context.Background())

	if len(remover.received_ids) != len(ids) {
		t.Fatalf("received %d ids, want %d", len(remover.received_ids), len(ids))
	}
	for i, id := range ids {
		if remover.received_ids[i] != id {
			t.Errorf("received_ids[%d] = %q, want %q", i, remover.received_ids[i], id)
		}
	}
}
