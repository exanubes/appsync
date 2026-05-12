package subscription

import (
	"context"
	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
)

type UnsubscribeService struct {
	subscriptions UnsubscribeRegistry
	writer        app.SendMessageService
	authorizer    app.RequestAuthorizer
	frame         FrameFactory
}

func NewUnsubscribeService(
	subscriptions UnsubscribeRegistry,
	writer app.SendMessageService,
	authorizer app.RequestAuthorizer,
	frame FrameFactory,
) *UnsubscribeService {
	return &UnsubscribeService{
		frame:         frame,
		authorizer:    authorizer,
		writer:        writer,
		subscriptions: subscriptions,
	}
}

func (service UnsubscribeService) Unsubscribe(ctx context.Context, subscription_id string) error {
	frame := service.frame.Unsubscribe()
	subscription := service.subscriptions.Get(subscription_id)

	if subscription == nil {
		return app.ErrSubscriptionClosed
	}

	if !subscription.Active() {
		return app.ErrSubscriptionClosed
	}

	signature, err := service.authorizer.Authorize(ctx, app.AuthorizeCommandInput{})
	if err != nil {
		return err
	}

	frame.WithType(protocol.TypeUnsubscribe).
		WithSignature(signature).
		WithID(subscription_id)

	err = service.writer.Send(ctx, frame.Build())

	if err != nil {
		return err
	}

	subscription.Close()
	service.subscriptions.Remove(subscription_id)

	return nil
}
