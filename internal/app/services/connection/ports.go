package connection

import (
	"context"
	"net/url"
	"time"

	"github.com/exanubes/appsync/internal/app"
)

type CreateConnectionInput struct {
	Url          *url.URL
	Subprotocols []string
}

type CreateConnectionOutput struct {
	Connection Connection
	Timeout    time.Duration
}

type DialOptions struct {
	Url          *url.URL
	Subprotocols []string
}

type Dialer interface {
	Dial(context.Context, DialOptions) (Connection, error)
}

type ConnectionAuthorizer interface {
	Authorize(context.Context, Connection) (time.Duration, error)
}

type SubprotocolGenerator interface {
	Generate(context.Context) (string, error)
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
	Close(context.Context) error
}

type Serializer interface {
	Serialize(app.Signature) (string, error)
}

type Runtime interface {
	Close(context.Context) error
}

type Transport interface {
	Close(context.Context) error
}
