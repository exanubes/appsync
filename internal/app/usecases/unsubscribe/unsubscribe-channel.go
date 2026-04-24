package unsubscribe

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
)

type UnsubscribeChannelUsecase struct {
	subscriptions Registry
	writer        app.SendMessageService
	authorizer    app.RequestAuthorizer
}

func NewUnsubscribeChannelUsecase(registry Registry, authorizer app.RequestAuthorizer, writer app.SendMessageService) *UnsubscribeChannelUsecase {
	return &UnsubscribeChannelUsecase{
		subscriptions: registry,
		authorizer:    authorizer,
		writer:        writer,
	}
}

func (usecase *UnsubscribeChannelUsecase) Execute(ctx context.Context, input UnsubscribeChannelCommandInput) error {
	subscription := usecase.subscriptions.Get(input.SubscriptionId)

	if subscription == nil {
		return app.ErrSubscriptionClosed
	}

	if subscription.Active() == false {
		return app.ErrSubscriptionClosed
	}

	signature, err := usecase.authorizer.Authorize(ctx, app.AuthorizeCommandInput{})
	if err != nil {
		return err
	}

	input.Frame.WithType(protocol.TypeUnsubscribe).
		WithSignature(signature).
		WithID(input.SubscriptionId)

	err = usecase.writer.Send(ctx, input.Frame.Build())

	if err != nil {
		return err
	}

	subscription.Close()
	usecase.subscriptions.Remove(input.SubscriptionId)

	return nil

}
