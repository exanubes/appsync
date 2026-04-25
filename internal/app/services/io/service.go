package io

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/services/connection"
)

type IOService struct {
	conn    connection.Connection
	decoder app.Decoder
	inbox   Inbox
	outbox  Outbox
}

func New(inbox Inbox, outbox Outbox, reader connection.Connection, codec app.Decoder) *IOService {
	return &IOService{
		conn:    reader,
		decoder: codec,
		inbox:   inbox,
		outbox:  outbox,
	}
}

func (service *IOService) Write(ctx context.Context) error {
	for {
		payload, err := service.outbox.Next(ctx)
		if err != nil {
			return err
		}
		err = service.conn.Write(ctx, payload)
		if err != nil {
			return err
		}
	}
}
func (service *IOService) Read(ctx context.Context) error {
	for {
		data, err := service.conn.Read(ctx)

		if err != nil {
			return err
		}
		msg, err := service.decoder.Decode(data)

		if err != nil {
			return err
		}

		err = service.inbox.Enqueue(ctx, msg)

		if err != nil {
			return err
		}
	}
}
