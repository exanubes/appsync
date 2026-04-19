package router

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
	"github.com/exanubes/appsync/internal/app/usecases"
)

type MessageHandler struct {
	publisher usecases.PublishMessage
}

func New(publisher usecases.PublishMessage) *MessageHandler {
	return &MessageHandler{
		publisher: publisher,
	}
}

func (handler *MessageHandler) Handle(ctx context.Context, msg app.Message) error {
	switch msg := msg.(type) {
	case protocol.PublishMessage:
		return handler.publisher.Publish(ctx, msg)
	}

	return nil
}
