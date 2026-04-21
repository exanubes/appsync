package publish

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type PublishMessageUsecase struct {
	authorizer app.RequestAuthorizer
	writer     app.SendMessageService
}

func NewPublishMessageUsecase(
	authorizer app.RequestAuthorizer,
	writer app.SendMessageService,
) *PublishMessageUsecase {
	return &PublishMessageUsecase{
		authorizer: authorizer,
		writer:     writer,
	}
}

func (usecase *PublishMessageUsecase) Publish(ctx context.Context, input PublishCommandInput) error {
	frame := input.Frame
	frame.
		WithPayload(input.Payload).
		WithChannel(input.Destination)

	signature, err := usecase.authorizer.Authorize(ctx, app.AuthorizeCommandInput{
		Channel: input.Destination,
		Payload: input.Payload,
	})

	if err != nil {
		return err
	}

	frame.WithSignature(signature)
	return usecase.writer.Send(ctx, frame.Build())
}
