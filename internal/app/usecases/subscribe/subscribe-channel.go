package subscribe

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/subscription"
)

// TODO: config that defines buffer sizes
var buffer_size uint = 100

type SubscribeChannelUseCase struct {
	authorizer app.RequestAuthorizer
	writer     app.SendMessageService
}

func NewSubscribeChannelUsecase(
	authorizer app.RequestAuthorizer,
	writer app.SendMessageService,
) *SubscribeChannelUseCase {
	return &SubscribeChannelUseCase{
		authorizer: authorizer,
		writer:     writer,
	}
}

func (usecase *SubscribeChannelUseCase) Execute(ctx context.Context, input SubscribeCommandInput) (*SubscribeCommandOutput, error) {
	frame := input.Frame
	frame.WithChannel(input.Channel)

	signature, err := usecase.authorizer.Authorize(ctx, app.AuthorizeCommandInput{
		Channel: input.Channel,
	})

	if err != nil {
		return nil, err
	}

	frame.WithSignature(signature)

	msg := frame.Build()

	err = usecase.writer.Send(ctx, msg)

	if err != nil {
		return nil, err
	}

	sub := subscription.New(msg.ID(), input.Channel, buffer_size)

	return &SubscribeCommandOutput{
		Subscription: sub,
	}, nil
}
