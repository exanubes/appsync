package composition

import (
	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/pending"
	"github.com/exanubes/appsync/internal/app/queue"
	"github.com/exanubes/appsync/internal/app/services/request"
	"github.com/exanubes/appsync/internal/app/usecases/publish"
)

type UseCases struct {
	Publish publish.PublishMessage
}

func NewUseCases(authorizer app.RequestAuthorizer,
	ingress *queue.IngressQueue,
	egress *queue.EgressQueue,
	pending *pending.Registry,
) *UseCases {
	send_request_service := request.NewSendRequestService(egress, pending)
	publish_usecase := publish.NewPublishMessageUsecase(authorizer, send_request_service)

	return &UseCases{
		Publish: publish_usecase,
	}
}
