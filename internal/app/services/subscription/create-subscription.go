package subscription

import (
	"github.com/exanubes/appsync/internal/app/subscription"
)

type CreateSubscriptionService struct {
	registry     SubscribeRegistry
	backpressure uint
}

func NewCreateSubscriptionService(registry SubscribeRegistry, backpressure uint) *CreateSubscriptionService {
	return &CreateSubscriptionService{
		registry:     registry,
		backpressure: backpressure,
	}
}

func (service *CreateSubscriptionService) Create(input CreateSubscriptionInput) (*subscription.Subscription, error) {
	sub, err := subscription.New(input.ID, input.Channel, service.backpressure)
	if err != nil {
		return nil, err
	}
	service.registry.Register(sub)
	return sub, err
}
