package appsync

import (
	"context"

	"github.com/exanubes/appsync/internal/app/usecases/publish"
	"github.com/exanubes/appsync/internal/app/usecases/subscribe"
	"github.com/exanubes/appsync/internal/composition"
	"github.com/exanubes/appsync/internal/infrastructure/events"
)

const (
	ProtocolEvents = "aws-appsync-event-ws"
)

type appsync_client struct {
	usecases *composition.UseCases
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
	return client.usecases.Shutdown.Execute(ctx)
}
