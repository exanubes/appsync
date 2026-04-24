package unsubscribe

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/subscription"
)

type UnsubscribeChannelCommandInput struct {
	SubscriptionId string
	Frame          app.FrameBuilder
}

type Registry interface {
	Remove(id string)
	Get(id string) *subscription.Subscription
}

type UnsubscribeChannel interface {
	Execute(context.Context, UnsubscribeChannelCommandInput) error
}
