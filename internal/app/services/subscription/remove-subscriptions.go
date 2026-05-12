package subscription

import (
	"context"
	"errors"
	"sync"
)

type RemoveSubscriptionsService struct {
	subscription Unsubscriber
}

func NewRemoveSubscriptionsService(unsubscribe Unsubscriber) *RemoveSubscriptionsService {
	return &RemoveSubscriptionsService{
		subscription: unsubscribe,
	}
}

func (service *RemoveSubscriptionsService) Remove(ctx context.Context, subscription_ids ...string) error {
	var wg sync.WaitGroup
	var err error
	error_channel := make(chan error, len(subscription_ids))

	for _, id := range subscription_ids {
		wg.Go(func() {
			select {
			case error_channel <- service.subscription.Unsubscribe(ctx, id):
			case <-ctx.Done():
			default:
			}
		})
	}
	wg.Wait()
	index := len(subscription_ids)
	for index > 0 {
		index -= 1
		select {
		case <-ctx.Done():
			return ctx.Err()
		case unsub_error := <-error_channel:
			err = errors.Join(err, unsub_error)
		default:
		}
	}

	return err
}
