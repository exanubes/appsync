package engine_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/engine"
)

type stub_logger struct {
	ctx string
}

func (s *stub_logger) Debug(string, ...any) {}
func (s *stub_logger) SetContext(c string) app.Logger {
	s.ctx = c
	return s
}

type mock_io struct {
	mu           sync.Mutex
	read_called  bool
	write_called bool
	err          error
}

func (m *mock_io) Read(ctx context.Context) error {
	m.mu.Lock()
	m.read_called = true
	m.mu.Unlock()
	<-ctx.Done()
	return m.err
}

func (m *mock_io) Write(ctx context.Context) error {
	m.mu.Lock()
	m.write_called = true
	m.mu.Unlock()
	<-ctx.Done()
	return m.err
}

type mock_runtime struct {
	called bool
	err    error
}

func (m *mock_runtime) Run(ctx context.Context) error {
	m.called = true
	<-ctx.Done()
	return m.err
}

type mock_heartbeat struct {
	called bool
	err    error
}

func (m *mock_heartbeat) Start(ctx context.Context, _ time.Duration) error {
	m.called = true
	<-ctx.Done()
	return m.err
}

func (m *mock_heartbeat) Reset() {}

func TestNew(t *testing.T) {
	logger := &stub_logger{}
	engine.New(&mock_heartbeat{}, &mock_runtime{}, &mock_io{}, logger)
	if logger.ctx != "Engine" {
		t.Errorf("expected logger context 'Engine', got %q", logger.ctx)
	}
}

func TestStart_calls_all_goroutines(t *testing.T) {
	io := &mock_io{}
	runtime := &mock_runtime{}
	hb := &mock_heartbeat{}

	e := engine.New(hb, runtime, io, &stub_logger{})
	e.Start(context.Background(), engine.StartEngineInput{Timeout: time.Second})
	e.Close(context.Background())

	io.mu.Lock()
	read_called := io.read_called
	write_called := io.write_called
	io.mu.Unlock()

	if !read_called {
		t.Error("expected io.Read to be called")
	}
	if !write_called {
		t.Error("expected io.Write to be called")
	}
	if !runtime.called {
		t.Error("expected runtime.Run to be called")
	}
	if !hb.called {
		t.Error("expected heartbeat.Start to be called")
	}
}

func TestClose(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name          string
		err           error
		want_err      bool
		want_err_wrap error
	}{
		{
			name:     "returns nil when goroutines return nil",
			err:      nil,
			want_err: false,
		},
		{
			name:     "returns nil when goroutines return context.Canceled",
			err:      context.Canceled,
			want_err: false,
		},
		{
			name:          "returns error when goroutines return non-canceled error",
			err:           boom,
			want_err:      true,
			want_err_wrap: boom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io := &mock_io{err: tt.err}
			runtime := &mock_runtime{err: tt.err}
			hb := &mock_heartbeat{err: tt.err}

			e := engine.New(hb, runtime, io, &stub_logger{})
			e.Start(context.Background(), engine.StartEngineInput{Timeout: time.Second})
			err := e.Close(context.Background())

			if tt.want_err && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.want_err && err != nil {
				t.Errorf("expected nil error, got %v", err)
			}
			if tt.want_err_wrap != nil && !errors.Is(err, tt.want_err_wrap) {
				t.Errorf("expected error to wrap %v, got %v", tt.want_err_wrap, err)
			}
		})
	}
}
