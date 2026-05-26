package appsync

import (
	"context"

	"github.com/exanubes/appsync/internal/app/lifecycle"
	"github.com/exanubes/appsync/internal/app/usecases/publish"
	"github.com/exanubes/appsync/internal/app/usecases/subscribe"
	"github.com/exanubes/appsync/internal/composition"
	"github.com/exanubes/appsync/internal/infrastructure/events"
)

// ProtocolEvents is the Appsync Events Websocket subprotocol
const ProtocolEvents = "aws-appsync-event-ws"

type appsync_client struct {
	usecases   *composition.UseCases
	connection *lifecycle.State
}

func (client *appsync_client) Publish(ctx context.Context, input PublishCommandInput) error {
	if err := client.connection.Err(); err != nil {
		return err
	}

	frame := &events.FrameBuilder{}

	err := client.usecases.Publish.Publish(ctx, publish.PublishCommandInput{
		Destination: input.Channel,
		Payload:     input.Payload,
		Frame:       frame,
	})
	return err
}

func (client *appsync_client) Subscribe(ctx context.Context, input SubscribeCommandInput) (Subscription, error) {
	if err := client.connection.Err(); err != nil {
		return nil, err
	}

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
	return client.usecases.Shutdown.Execute(ctx)
}
