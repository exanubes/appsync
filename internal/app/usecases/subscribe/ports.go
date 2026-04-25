package subscribe

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	sub_service "github.com/exanubes/appsync/internal/app/services/subscription"
	"github.com/exanubes/appsync/internal/app/subscription"
)

type SubscribeCommandInput struct {
	Channel string
	Frame   app.FrameBuilder
}
type SubscribeCommandOutput struct {
	SubID        string
	Subscription *subscription.Subscription
}

type SubscribeChannel interface {
	Execute(context.Context, SubscribeCommandInput) (*SubscribeCommandOutput, error)
}

type CreateSubscription interface {
	Create(sub_service.CreateSubscriptionInput) (*subscription.Subscription, error)
}
