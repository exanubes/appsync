package usecases

import (
	"context"

	"github.com/exanubes/appsync/internal/app/protocol"
)

type PublishMessage interface {
	Publish(context.Context, protocol.PublishMessage) error
}
