package shutdown

import (
	"context"
	"errors"
)

type ShutdownConnectionUsecase struct {
	subscriptions SubscriptionRegistry
	subscription  Remover
	runtime       Closer
	transport     Closer
}

func NewShutdownConnectionUsecase(
	registry SubscriptionRegistry,
	subscription Remover,
	runtime Closer,
	transport Closer,
) *ShutdownConnectionUsecase {
	return &ShutdownConnectionUsecase{
		subscriptions: registry,
		subscription:  subscription,
		runtime:       runtime,
		transport:     transport,
	}
}

func (usecase *ShutdownConnectionUsecase) Execute(ctx context.Context) error {
	active_subscription_ids := usecase.subscriptions.Active()
	unsub_err := usecase.subscription.Remove(ctx, active_subscription_ids...)
	conn_err := usecase.runtime.Close(ctx)
	trans_err := usecase.transport.Close(ctx)

	return errors.Join(unsub_err, conn_err, trans_err)

}
