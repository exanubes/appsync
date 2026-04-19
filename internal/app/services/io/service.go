package io

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/queue"
	"github.com/exanubes/appsync/internal/app/services/connection"
)

type IOService struct {
	conn    connection.Connection
	decoder app.Decoder
}

func New(reader connection.Connection, codec app.Decoder) *IOService {
	return &IOService{
		conn:    reader,
		decoder: codec,
	}
}

func (service *IOService) Write(ctx context.Context, queue *queue.EgressQueue) error {
	for {
		payload, err := queue.Next(ctx)
		if err != nil {
			return err
		}
		err = service.conn.Write(ctx, payload)
		if err != nil {
			return err
		}
	}
}
func (service *IOService) Read(ctx context.Context, queue *queue.IngressQueue) error {
	for {
		data, err := service.conn.Read(ctx)

		if err != nil {
			return err
		}
		msg, err := service.decoder.Decode(data)

		if err != nil {
			return err
		}

		err = queue.Enqueue(ctx, msg)

		if err != nil {
			return err
		}
	}
}
