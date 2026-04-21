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
	return nil, nil
}

func (sub *ChannelSubscription) Decode(ctx context.Context, value any) error {
	return nil
}
