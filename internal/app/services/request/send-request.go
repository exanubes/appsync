package request

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/pending"
	"github.com/exanubes/appsync/internal/app/queue"
)

type SendRequestService struct {
	egress  *queue.EgressQueue
	pending *pending.Registry
}

func NewSendRequestService(
	egress *queue.EgressQueue,
	pending *pending.Registry,

) *SendRequestService {
	return &SendRequestService{
		egress:  egress,
		pending: pending,
	}
}

func (service *SendRequestService) Send(ctx context.Context, input app.Frame) error {
	if service.pending.Has(input.ID()) {
		return app.ErrDuplicateMessage
	}
	payload, err := input.Encode()
	if err != nil {
		return err
	}
	err = service.egress.Enqueue(ctx, payload)

	if err != nil {
		return err
	}
	service.pending.Register(input.ID())

	return service.pending.Consume(ctx, input.ID())
}
