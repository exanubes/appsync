package shutdown

import "context"

type ShutdownConnection interface {
	Execute(context.Context) error
}

type SubscriptionRegistry interface {
	Active() []string
}

type Closer interface {
	Close(context.Context) error
}

type Remover interface {
	Remove(context.Context, ...string) error
}

type Runtime interface {
	Close(context.Context) error
}

type Transport interface {
	Close(context.Context) error
}
