package subscribe

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
	"github.com/exanubes/appsync/internal/app/services/subscription"
)

type SubscribeChannelUseCase struct {
	authorizer   app.RequestAuthorizer
	writer       app.SendMessageService
	subscription CreateSubscription
}

func NewSubscribeChannelUsecase(
	authorizer app.RequestAuthorizer,
	writer app.SendMessageService,
	subscription CreateSubscription,
) *SubscribeChannelUseCase {
	return &SubscribeChannelUseCase{
		authorizer:   authorizer,
		writer:       writer,
		subscription: subscription,
	}
}

func (usecase *SubscribeChannelUseCase) Execute(ctx context.Context, input SubscribeCommandInput) (*SubscribeCommandOutput, error) {

	signature, err := usecase.authorizer.Authorize(ctx, app.AuthorizeCommandInput{
		Channel: input.Channel,
	})

	if err != nil {
		return nil, err
	}

	input.Frame.WithType(protocol.TypeSubscribe).
		WithChannel(input.Channel).
		WithSignature(signature)

	msg := input.Frame.Build()

	err = usecase.writer.Send(ctx, msg)

	if err != nil {
		return nil, err
	}

	sub := usecase.subscription.Create(subscription.CreateSubscriptionInput{
		ID:      msg.ID(),
		Channel: input.Channel,
	})

	return &SubscribeCommandOutput{
		SubID:        msg.ID(),
		Subscription: sub,
	}, nil
}
