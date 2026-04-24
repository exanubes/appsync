package router

import (
	"context"
	"fmt"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/pending"
	"github.com/exanubes/appsync/internal/app/protocol"
	"github.com/exanubes/appsync/internal/app/usecases/data"
)

type MessageHandler struct {
	pending          *pending.Registry
	receive_use_case data.ReceiveData
}

func New(pending *pending.Registry, receive data.ReceiveData) *MessageHandler {
	return &MessageHandler{
		pending:          pending,
		receive_use_case: receive,
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

	case protocol.DataMessage:
		return router.receive_use_case.Execute(ctx, msg)
	}

	return nil
}
