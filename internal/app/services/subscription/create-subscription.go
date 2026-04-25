package subscription

import (
	"github.com/exanubes/appsync/internal/app/subscription"
)

// TODO: config that defines buffer sizes
var buffer_size uint = 100

type CreateSubscriptionService struct {
	registry Registry
}

func NewCreateSubscriptionService(registry Registry) *CreateSubscriptionService {
	return &CreateSubscriptionService{
		registry: registry,
	}
}

func (service *CreateSubscriptionService) Create(input CreateSubscriptionInput) (*subscription.Subscription, error) {
	sub, err := subscription.New(input.ID, input.Channel, buffer_size)
	if err != nil {
		return nil, err
	}
	service.registry.Register(sub)
	return sub, err
}
