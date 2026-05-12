package events

import (
	"encoding/json"

	"github.com/exanubes/appsync/internal/app"
	"github.com/google/uuid"
)

type FrameBuilderFactory struct{}

func (FrameBuilderFactory) Create() app.FrameBuilder {
	return &FrameBuilder{}
}

func (FrameBuilderFactory) Unsubscribe() app.FrameBuilder {
	return &FrameBuilder{
		_type: "unsubscribe",
	}
}

type FrameBuilder struct {
	id        string
	_type     string
	payload   app.Payload
	channel   string
	signature app.Signature
}

func (builder *FrameBuilder) WithPayload(payload app.Payload) app.FrameBuilder {
	builder.payload = payload
	return builder
}

func (builder *FrameBuilder) WithChannel(channel string) app.FrameBuilder {
	builder.channel = channel
	return builder
}

func (builder *FrameBuilder) WithSignature(signature app.Signature) app.FrameBuilder {
	builder.signature = signature
	return builder
}

func (builder *FrameBuilder) WithType(typ string) app.FrameBuilder {
	builder._type = typ
	return builder
}

func (builder *FrameBuilder) WithID(id string) app.FrameBuilder {
	builder.id = id
	return builder
}

func (builder *FrameBuilder) Build() app.Frame {
	var id string

	if builder.id != "" {
		id = builder.id
	} else {
		id = uuid.NewString()
	}
	return Frame{
		Id:        id,
		Topic:     builder.channel,
		Signature: builder.signature,
		Payload:   []string{string(builder.payload)},
		Type:      builder._type,
	}
}

type Frame struct {
	Type      string        `json:"type,omitempty"`
	Id        string        `json:"id,omitempty"`
	Topic     string        `json:"channel,omitempty"`
	Payload   []string      `json:"events,omitempty"`
	Signature app.Signature `json:"authorization,omitempty"`
}

func (f Frame) Encode() (app.Payload, error) {
	return json.Marshal(f)
}

func (f Frame) ID() string {
	return f.Id
}
