package publish

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
)

type PublishCommandInput struct {
	Destination string
	Payload     app.Payload
	Frame       app.FrameBuilder
}

type PublishMessage interface {
	Publish(context.Context, PublishCommandInput) error
}

type ReceivePublishResult interface {
	Receive(context.Context, protocol.PublishResult)
}

type Reply chan error

type SendMessageService interface {
	Send(context.Context, app.Frame) error
}
