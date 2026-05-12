package appsync

import (
	"context"

	"github.com/exanubes/appsync/internal/app/subscription"
	"github.com/exanubes/appsync/internal/app/usecases/unsubscribe"
)

type channel_subscription struct {
	id           string
	subscription *subscription.Subscription
	unsubscribe  unsubscribe.UnsubscribeChannel
}

func (sub *channel_subscription) Close(ctx context.Context) error {
	return sub.unsubscribe.Execute(ctx, unsubscribe.UnsubscribeChannelCommandInput{
		SubscriptionId: sub.id,
	})
}

func (sub *channel_subscription) Next(ctx context.Context) (*NextMessageOutput, error) {
	data, err := sub.subscription.Next(ctx)

	if err != nil {
		return nil, err
	}
	return &NextMessageOutput{
		Data: data,
	}, nil
}

func (sub *channel_subscription) DecodeNext(ctx context.Context, value any) error {
	return sub.subscription.Decode(ctx, value)
}
