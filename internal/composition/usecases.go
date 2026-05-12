package composition

import (
	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/pending"
	"github.com/exanubes/appsync/internal/app/queue"
	"github.com/exanubes/appsync/internal/app/services/request"
	sub_service "github.com/exanubes/appsync/internal/app/services/subscription"
	"github.com/exanubes/appsync/internal/app/subscription"
	"github.com/exanubes/appsync/internal/app/usecases/data"
	"github.com/exanubes/appsync/internal/app/usecases/publish"
	"github.com/exanubes/appsync/internal/app/usecases/shutdown"
	"github.com/exanubes/appsync/internal/app/usecases/subscribe"
	"github.com/exanubes/appsync/internal/app/usecases/unsubscribe"
	"github.com/exanubes/appsync/internal/infrastructure/events"
)

type UseCases struct {
	Publish     publish.PublishMessage
	Subscribe   subscribe.SubscribeChannel
	ReceiveData data.ReceiveData
	Unsubscribe unsubscribe.UnsubscribeChannel
	Shutdown    shutdown.ShutdownConnection
}

type Services struct {
	SubscriptionsRegistry *subscription.Registry
	RemoveSubscriptions   *sub_service.RemoveSubscriptionsService
}

func NewUseCases(authorizer app.RequestAuthorizer,
	ingress *queue.IngressQueue,
	egress *queue.EgressQueue,
	pending *pending.Registry,
	subscription_backpressure uint,
) (*UseCases, *Services) {
	subscriptions_registry := subscription.NewRegistry()
	create_subscription_service := sub_service.NewCreateSubscriptionService(subscriptions_registry, subscription_backpressure)
	send_request_service := request.NewSendRequestService(egress, pending)
	unsubscribe_service := sub_service.NewUnsubscribeService(subscriptions_registry,
		send_request_service,
		authorizer,
		events.FrameBuilderFactory{},
	)

	remove_subscriptions_service := sub_service.NewRemoveSubscriptionsService(
		unsubscribe_service,
	)

	publish_usecase := publish.NewPublishMessageUsecase(authorizer, send_request_service)
	subscribe_usecase := subscribe.NewSubscribeChannelUsecase(authorizer, send_request_service, create_subscription_service)
	unsubscribe_usecase := unsubscribe.NewUnsubscribeChannelUsecase(unsubscribe_service)
	receive_data_usecase := data.NewReceiveDataUsecase(subscriptions_registry)
	return &UseCases{
			Publish:     publish_usecase,
			Subscribe:   subscribe_usecase,
			ReceiveData: receive_data_usecase,
			Unsubscribe: unsubscribe_usecase,
		}, &Services{
			SubscriptionsRegistry: subscriptions_registry,
			RemoveSubscriptions:   remove_subscriptions_service,
		}
}
