package subscription

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/subscription"
)

type SubscribeRegistry interface {
	Register(*subscription.Subscription)
}

type UnsubscribeRegistry interface {
	Remove(id string)
	Get(id string) *subscription.Subscription
}

type CreateSubscriptionInput struct {
	ID      string
	Channel string
}

type Subscriptions interface {
	Active() []string
	Get(string) *subscription.Subscription
}

type FrameFactory interface {
	Unsubscribe() app.FrameBuilder
}

type Unsubscriber interface {
	Unsubscribe(ctx context.Context, subscription_id string) error
}
