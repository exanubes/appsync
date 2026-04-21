package app

import (
	"context"

	"github.com/exanubes/appsync/internal/app/subscription"
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
	Encode(Frame) (Payload, error)
}

type Decoder interface {
	Decode(Payload) (Message, error)
}

// INFO: Temporary placeholder type. Final structure TBD
type Message any

type Signature map[string]string
type Payload []byte

type AuthorizeCommandInput struct {
	Channel string
	Payload []byte
}

type RequestAuthorizer interface {
	Authorize(context.Context, AuthorizeCommandInput) (Signature, error)
}

type Heartbeat interface {
	Start(context.Context) <-chan error
	Reset()
}

type Router interface {
	Handle(context.Context, Message) error
}

type Inbox interface {
	Next(context.Context) (Message, error)
}

type FrameBuilder interface {
	WithPayload(Payload) FrameBuilder
	WithChannel(string) FrameBuilder
	WithSignature(Signature) FrameBuilder
	Build() Frame
}

type FrameBuilderFactory interface {
	Create() FrameBuilder
}

type Frame interface {
	ID() string
	Encode() (Payload, error)
}

type SendMessageService interface {
	Send(context.Context, Frame) error
}
