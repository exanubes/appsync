package unsubscribe

import (
	"context"

	"github.com/exanubes/appsync/internal/app/subscription"
)

type UnsubscribeChannelCommandInput struct {
	SubscriptionId string
}

type Registry interface {
	Remove(id string)
	Get(id string) *subscription.Subscription
}

type Unsubscriber interface {
	Unsubscribe(context.Context, string) error
}

type UnsubscribeChannel interface {
	Execute(context.Context, UnsubscribeChannelCommandInput) error
}
