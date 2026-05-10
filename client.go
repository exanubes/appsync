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
)

type appsync_client struct {
	transport connection.Connection
	runtime   *engine.Engine
	usecases  *composition.UseCases
}

func (client *appsync_client) Publish(ctx context.Context, input PublishCommandInput) error {
	frame := &events.FrameBuilder{}

	err := client.usecases.Publish.Publish(ctx, publish.PublishCommandInput{
		Destination: input.Channel,
		Payload:     input.Payload,
		Frame:       frame,
	})
	return err
}

func (client *appsync_client) Subscribe(ctx context.Context, input SubscribeCommandInput) (Subscription, error) {
	frame := &events.FrameBuilder{}
	result, err := client.usecases.Subscribe.Execute(ctx, subscribe.SubscribeCommandInput{
		Channel: input.Channel,
		Frame:   frame,
	})
	if err != nil {
		return nil, err
	}
	return &channel_subscription{
		id:           result.SubID,
		subscription: result.Subscription,
		unsubscribe:  client.usecases.Unsubscribe,
	}, nil
}

func (client *appsync_client) Close(ctx context.Context) error {
	client.runtime.Close(ctx)
	return client.transport.Close()
}
