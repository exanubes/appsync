package connection

import (
	"context"
	"time"
)

type CreateConnectionInput struct {
	Url          string
	Subprotocols []string
}
type CreateConnectionOutput struct {
	Connection Connection
	Timeout    time.Duration
}

type CreateConnectionService interface {
	Connect(context.Context, CreateConnectionInput) (*CreateConnectionOutput, error)
}

type DialOptions struct {
	Url          string
	Subprotocols []string
}

type Dialer interface {
	Dial(context.Context, DialOptions) (Connection, error)
}

type ConnectionAuthorizer interface {
	Authorize(context.Context, Connection) (time.Duration, error)
}

type Connection interface {
	Reader
	Writer
	Closer
}

type Reader interface {
	Read(context.Context) ([]byte, error)
}

type Writer interface {
	Write(context.Context, []byte) error
}

type Closer interface {
	Close() error
}
