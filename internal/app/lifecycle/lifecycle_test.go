package lifecycle_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/lifecycle"
)

var errCustom = errors.New("custom error")

func closed_state(cause error) *lifecycle.State {
	s := lifecycle.NewState()
	s.Close(cause)
	return s
}

func TestNewState(t *testing.T) {
	t.Run("done channel is open", func(t *testing.T) {
		s := lifecycle.NewState()
		select {
		case <-s.Done():
			t.Fatal("Done() should not be readable on a new state")
		default:
		}
	})
}

func TestDone(t *testing.T) {
	t.Run("blocks before Close", func(t *testing.T) {
		s := lifecycle.NewState()
		select {
		case <-s.Done():
			t.Fatal("Done() should block before Close()")
		default:
		}
	})

	t.Run("readable after Close", func(t *testing.T) {
		s := closed_state(nil)
		select {
		case <-s.Done():
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Done() should be readable after Close()")
		}
	})
}

func TestErr(t *testing.T) {
	tests := []struct {
		name      string
		state     func() *lifecycle.State
		is_closed bool
		is_custom bool
	}{
		{
			name:      "returns nil before Close",
			state:     lifecycle.NewState,
			is_closed: false,
			is_custom: false,
		},
		{
			name:      "returns ErrConnectionClosed when closed with nil",
			state:     func() *lifecycle.State { return closed_state(nil) },
			is_closed: true,
			is_custom: false,
		},
		{
			name:      "returns ErrConnectionClosed when closed with ErrConnectionClosed",
			state:     func() *lifecycle.State { return closed_state(app.ErrConnectionClosed) },
			is_closed: true,
			is_custom: false,
		},
		{
			name:      "wraps custom error with ErrConnectionClosed",
			state:     func() *lifecycle.State { return closed_state(errCustom) },
			is_closed: true,
			is_custom: true,
		},
		{
			name:      "preserves pre-joined error containing ErrConnectionClosed and custom error",
			state:     func() *lifecycle.State { return closed_state(errors.Join(app.ErrConnectionClosed, errCustom)) },
			is_closed: true,
			is_custom: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.state().Err()

			if !tt.is_closed {
				if err != nil {
					t.Errorf("Err() = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Fatal("Err() = nil, want non-nil error")
			}
			if !errors.Is(err, app.ErrConnectionClosed) {
				t.Errorf("Err() = %v, want errors.Is(err, ErrConnectionClosed) == true", err)
			}
			if tt.is_custom && !errors.Is(err, errCustom) {
				t.Errorf("Err() = %v, want errors.Is(err, errCustom) == true", err)
			}
		})
	}
}

func TestClose(t *testing.T) {
	t.Run("closes Done channel", func(t *testing.T) {
		s := lifecycle.NewState()
		s.Close(nil)
		select {
		case <-s.Done():
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Done() should be closed after Close()")
		}
	})

	t.Run("idempotent: second call does not overwrite first", func(t *testing.T) {
		s := lifecycle.NewState()
		s.Close(errCustom)
		s.Close(nil)

		if err := s.Err(); !errors.Is(err, errCustom) {
			t.Errorf("Err() = %v, want errors.Is(err, errCustom) == true after second Close(nil)", err)
		}
	})

	t.Run("concurrent calls are safe and close exactly once", func(t *testing.T) {
		const goroutines = 100
		s := lifecycle.NewState()

		var wg sync.WaitGroup
		wg.Add(goroutines)
		start := make(chan struct{})

		for range goroutines {
			go func() {
				defer wg.Done()
				<-start
				s.Close(errCustom)
			}()
		}

		close(start)
		wg.Wait()

		if err := s.Err(); err == nil {
			t.Fatal("Err() = nil after concurrent Close calls, want non-nil")
		}
	})
}
