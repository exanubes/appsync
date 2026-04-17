package app

import (
	"context"
)

type DialOptions struct {
	Url          string
	Subprotocols []string
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

type Connection interface {
	Reader
	Writer
	Closer
}

type Logger interface {
	Debug(string, ...any)
	SetContext(string) Logger
}

type Codec interface {
	Encoder
	Decoder
}

type Encoder interface {
	Encode(Message, Signature) (Payload, error)
}

type Decoder interface {
	Decode(Payload) (Message, error)
}

// INFO: Temporary placeholder type. Final structure TBD
type Message any

type Signature map[string]string
type Payload []byte

type Authorizer interface {
	Authorize(context.Context, Payload) (Signature, error)
}

type Heartbeat interface {
	Start(context.Context) <-chan error
	Reset()
}

type Engine interface {
	Start(context.Context) error
}
