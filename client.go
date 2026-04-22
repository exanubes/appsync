package appsync

import (
	"context"

	"github.com/exanubes/appsync/internal/app/engine"
	"github.com/exanubes/appsync/internal/app/services/connection"
	"github.com/exanubes/appsync/internal/app/usecases/publish"
	"github.com/exanubes/appsync/internal/app/usecases/subscribe"
	"github.com/exanubes/appsync/internal/composition"
	"github.com/exanubes/appsync/internal/infrastructure/events"
)

const (
	ProtocolEvents = "aws-appsync-event-ws"
	// ProtocolGraphql = "graphql-ws"
)

type AppsyncClient struct {
	transport connection.Connection
	runtime   *engine.Engine
	usecases  *composition.UseCases
}

func (client *AppsyncClient) Publish(ctx context.Context, input PublishCommandInput) error {
	frame := &events.FrameBuilder{}

	err := client.usecases.Publish.Publish(ctx, publish.PublishCommandInput{
		Destination: input.Channel,
		Payload:     input.Payload,
		Frame:       frame,
	})
	return err
}

func (client *AppsyncClient) Subscribe(ctx context.Context, input SubscribeCommandInput) (*SubscribeCommandOutput, error) {
	frame := &events.FrameBuilder{}
	result, err := client.usecases.Subscribe.Execute(ctx, subscribe.SubscribeCommandInput{
		Channel: input.Channel,
		Frame:   frame,
	})
	if err != nil {
		return nil, err
	}
	return &SubscribeCommandOutput{
		Sub: &ChannelSubscription{subscription: result.Subscription},
	}, err
}

func (client *AppsyncClient) Close(ctx context.Context) error {
	client.runtime.Close(ctx)
	return client.transport.Close()
}
