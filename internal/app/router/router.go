package router

import (
	"context"
	"fmt"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/pending"
	"github.com/exanubes/appsync/internal/app/protocol"
)

type MessageHandler struct {
	pending *pending.Registry
}

func New(pending *pending.Registry) *MessageHandler {
	return &MessageHandler{
		pending: pending,
	}
}

func (router *MessageHandler) Handle(ctx context.Context, msg app.Message) error {
	switch msg := msg.(type) {
	case protocol.ErrorMessage:
		if msg.ID == "" {
			return fmt.Errorf("%+v", msg.Errors)
		}

		return router.pending.Fulfill(ctx, msg.ID, fmt.Errorf("%+v", msg.Errors))
	case protocol.SuccessMessage:
		return router.pending.Fulfill(ctx, msg.ID, nil)
	}

	return nil
}
