package request

import (
	"context"
)

type Outbox interface {
	Enqueue(context.Context, []byte) error
}

type Registry interface {
	Register(id string)
	Remove(id string)
	Consume(ctx context.Context, id string) error
	Has(id string) bool
}
