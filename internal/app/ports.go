package app

import (
	"context"
)

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

type RequestAuthorizer interface {
	Authorize(context.Context, Payload) (Signature, error)
}

type Heartbeat interface {
	Start(context.Context) <-chan error
	Reset()
}

type Engine interface {
	Start(context.Context) error
}
