package usecases

import (
	"context"

	"github.com/exanubes/appsync/internal/app/protocol"
)

type PublishMessageUsecase struct{}

func NewPublishMessageUsecase() *PublishMessageUsecase {
	return &PublishMessageUsecase{}
}

func (usecase *PublishMessageUsecase) Publish(ctx context.Context, msg protocol.PublishMessage) error {
	return nil
}
