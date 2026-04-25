package io

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type Outbox interface {
	Next(context.Context) ([]byte, error)
}

type Inbox interface {
	Enqueue(context.Context, app.Message) error
}
