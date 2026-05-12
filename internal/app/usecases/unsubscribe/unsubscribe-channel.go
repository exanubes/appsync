package unsubscribe

import (
	"context"
)

type UnsubscribeChannelUsecase struct {
	subscription Unsubscriber
}

func NewUnsubscribeChannelUsecase(unsubscriber Unsubscriber) *UnsubscribeChannelUsecase {
	return &UnsubscribeChannelUsecase{
		subscription: unsubscriber,
	}
}

func (usecase *UnsubscribeChannelUsecase) Execute(ctx context.Context, input UnsubscribeChannelCommandInput) error {
	return usecase.subscription.Unsubscribe(ctx, input.SubscriptionId)
}
