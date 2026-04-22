package subscription

import (
	"github.com/exanubes/appsync/internal/app/subscription"
)

type Registry interface {
	Register(*subscription.Subscription)
}

type CreateSubscriptionInput struct {
	ID      string
	Channel string
}
