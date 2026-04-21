package subscribe

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/subscription"
)

type SubscribeCommandInput struct {
	Channel string
	Frame   app.FrameBuilder
}
type SubscribeCommandOutput struct {
	Subscription *subscription.Subscription
}

type SubscribeChannel interface {
	Execute(context.Context, SubscribeCommandInput) (*SubscribeCommandOutput, error)
}
