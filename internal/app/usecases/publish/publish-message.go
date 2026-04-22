package publish

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
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
	signature, err := usecase.authorizer.Authorize(ctx, app.AuthorizeCommandInput{
		Channel: input.Destination,
		Payload: input.Payload,
	})

	if err != nil {
		return err
	}

	input.Frame.
		WithType(protocol.TypePublish).
		WithPayload(input.Payload).
		WithChannel(input.Destination).
		WithSignature(signature)

	return usecase.writer.Send(ctx, input.Frame.Build())
}
