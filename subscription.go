package appsync

import (
	"context"

	"github.com/exanubes/appsync/internal/app/subscription"
)

type ChannelSubscription struct {
	subscription *subscription.Subscription
}

func (sub *ChannelSubscription) Close(ctx context.Context) error {
	return nil
}

func (sub *ChannelSubscription) Next(ctx context.Context) (*NextMessageOutput, error) {
	data, err := sub.subscription.Next(ctx)

	if err != nil {
		return nil, err
	}
	return &NextMessageOutput{
		Data: data,
	}, nil
}

func (sub *ChannelSubscription) DecodeNext(ctx context.Context, value any) error {
	return sub.subscription.Decode(ctx, value)
}
