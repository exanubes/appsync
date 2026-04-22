package data

import (
	"context"

	"github.com/exanubes/appsync/internal/app/protocol"
	"github.com/exanubes/appsync/internal/app/subscription"
)

type ReceiveData interface {
	Execute(context.Context, protocol.DataMessage) error
}

type Registry interface {
	Get(string) *subscription.Subscription
}
