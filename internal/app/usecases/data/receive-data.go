package data

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
)

type ReceiveDataUseCase struct {
	registry Registry
}

func NewReceiveDataUsecase(registry Registry) *ReceiveDataUseCase {
	return &ReceiveDataUseCase{registry}
}

func (usecase *ReceiveDataUseCase) Execute(ctx context.Context, msg protocol.DataMessage) error {
	subscription := usecase.registry.Get(msg.SubId)
	if subscription == nil {
		return app.ErrSubscriptionNotFound
	}

	return subscription.Deliver(ctx, msg.Payload)
}
