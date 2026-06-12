package publish

import (
	"context"
	"errors"
	"fmt"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
)

type PublishMessageUsecase struct {
	authorizer app.RequestAuthorizer
	writer     app.SendMessageService
	batcher    app.Batcher
}

func NewPublishMessageUsecase(
	authorizer app.RequestAuthorizer,
	writer app.SendMessageService,
	batcher app.Batcher,
) *PublishMessageUsecase {
	return &PublishMessageUsecase{
		authorizer: authorizer,
		writer:     writer,
		batcher:    batcher,
	}
}

func (usecase *PublishMessageUsecase) Publish(ctx context.Context, input PublishCommandInput) (*PublishCommandOutput, error) {
	authorizer := input.resolve_authorizer(usecase.authorizer)

	if authorizer == nil {
		return nil, app.ErrPublishAuthorizerMissing
	}

	batches := usecase.batcher.Batch(input.Events)
	failed_events := make([]FailedEvent, 0)
	success := true
	for _, batch := range batches {
		signature, err := authorizer.Authorize(ctx, app.AuthorizeCommandInput{
			Channel: input.Destination,
			Payload: batch,
		})

		if err != nil {
			return nil, err
		}

		input.Frame.
			WithType(protocol.TypePublish).
			WithBatch(batch).
			WithChannel(input.Destination).
			WithSignature(signature)

		err = usecase.writer.Send(ctx, input.Frame.Build())

		if err != nil {

			var batch_err protocol.BatchPublishError
			if errors.As(err, &batch_err) {
				success = false
				for _, failure := range batch_err.Failures {
					failed_events = append(failed_events, FailedEvent{
						Payload: batch[failure.Index],
						Err:     fmt.Errorf("Event failed to publish"),
					})
				}
			} else {
				return nil, err
			}

		}
	}

	return &PublishCommandOutput{
		Failures: failed_events,
		Success:  success,
	}, nil
}
