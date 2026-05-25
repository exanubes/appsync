package lifecycle

import (
	"errors"
	"sync"

	"github.com/exanubes/appsync/internal/app"
)

type State struct {
	once  sync.Once
	done  chan struct{}
	mutex sync.RWMutex
	err   error
}

func NewState() *State {
	return &State{
		done: make(chan struct{}),
	}
}

func (state *State) Done() <-chan struct{} {
	return state.done
}

func (state *State) Err() error {
	select {
	case <-state.done:
		state.mutex.RLock()
		defer state.mutex.RUnlock()

		if state.err == nil {
			return app.ErrConnectionClosed
		}

		return state.err
	default:
		return nil
	}
}

func (state *State) Close(cause error) {
	state.once.Do(func() {
		if cause == nil {
			cause = app.ErrConnectionClosed
		}

		if !errors.Is(cause, app.ErrConnectionClosed) {
			cause = errors.Join(app.ErrConnectionClosed, cause)
		}

		state.mutex.Lock()
		state.err = cause
		state.mutex.Unlock()

		close(state.done)
	})
}
