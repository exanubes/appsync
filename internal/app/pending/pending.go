package pending

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type Registry struct {
	store      map[string]chan error
	connection ConnectionState
}

func NewRegistry(connection ConnectionState) *Registry {
	return &Registry{
		store:      make(map[string]chan error),
		connection: connection,
	}
}

func (registry Registry) Has(id string) bool {
	_, exists := registry.store[id]
	return exists
}

func (registry Registry) Register(id string) {
	registry.store[id] = make(chan error, 1)
}

func (registry Registry) Fulfill(ctx context.Context, id string, err error) error {
	reply := registry.get(id)

	if reply == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case reply <- err:
		return nil
	case <-registry.connection.Done():
		return app.ErrConnectionClosed
	}
}

func (registry Registry) Remove(id string) {
	delete(registry.store, id)
}

func (registry Registry) get(id string) chan error {
	return registry.store[id]
}

func (registry Registry) Consume(ctx context.Context, id string) error {
	reply := registry.get(id)

	if reply == nil {
		return nil
	}

	defer delete(registry.store, id)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case res := <-reply:
		return res
	case <-registry.connection.Done():
		return app.ErrConnectionClosed
	}
}
