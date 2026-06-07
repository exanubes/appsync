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
	Authorizer  app.RequestAuthorizer
}

func (input *PublishCommandInput) resolve_authorizer(authz app.RequestAuthorizer) app.RequestAuthorizer {
	authorizer := authz

	if input.Authorizer != nil {
		authorizer = input.Authorizer
	}

	return authorizer
}

type PublishMessage interface {
	Publish(context.Context, PublishCommandInput) error
}

type ReceivePublishResult interface {
	Receive(context.Context, protocol.PublishResult)
}

type Reply chan error
