package appsync

import (
	"context"
	"net/url"

	"github.com/exanubes/appsync/internal/app/engine"
	"github.com/exanubes/appsync/internal/app/queue"
	"github.com/exanubes/appsync/internal/app/router"
	"github.com/exanubes/appsync/internal/app/runtime"
	"github.com/exanubes/appsync/internal/app/services/connection"
	"github.com/exanubes/appsync/internal/app/services/io"
	"github.com/exanubes/appsync/internal/app/usecases"
	"github.com/exanubes/appsync/internal/composition"
	"github.com/exanubes/appsync/internal/infrastructure/codec"
	"github.com/exanubes/appsync/internal/infrastructure/logger"
	"github.com/exanubes/appsync/internal/infrastructure/serializer"
	"github.com/exanubes/appsync/internal/infrastructure/transport"
)

const (
	ProtocolEvents  = "aws-appsync-event-ws"
	ProtocolGraphql = "graphql-ws"
)

// Connect establishes a new AppSync WebSocket connection and returns a Client.
func Connect(ctx context.Context, options ConnectionOptions) (*AppsyncClient, error) {
	http_endpoint, err := url.Parse(options.HttpEndpoint)

	if err != nil {
		return nil, err
	}

	ws_endpoint, err := url.Parse(options.WsEndpoint)

	if err != nil {
		return nil, err
	}

	request_authorizer := composition.NewIAMAuthorizer(options.Region, http_endpoint)
	slogger := logger.New()
	dialer := transport.New()
	msg_codec := codec.New()
	base64_serializer := serializer.New()

	generate_subprotocol_service := connection.NewGenerateSubprotocolService(request_authorizer, base64_serializer)
	authorize_connection_service := connection.NewAuthorizeConnectionService(msg_codec, request_authorizer, slogger)
	create_connection_service := connection.NewConnectionService(dialer, authorize_connection_service, generate_subprotocol_service, slogger)

	connection_output, err := create_connection_service.Connect(ctx, connection.CreateConnectionInput{
		Url:          ws_endpoint,
		Subprotocols: options.Subprotocols,
	})

	if err != nil {
		return nil, err
	}

	ingress_queue := queue.NewIngressQueue(100)
	egress_queue := queue.NewEgressQueue(100)
	io_loops := io.New(connection_output.Connection, msg_codec)

	publish_usecase := usecases.NewPublishMessageUsecase()

	msg_router := router.New(publish_usecase)
	runtime := runtime.New(msg_router)
	session := engine.New(runtime, io_loops, slogger)
	session.Start(ctx, engine.StartEngineInput{
		Timeout: connection_output.Timeout,
		Ingress: ingress_queue,
		Egress:  egress_queue,
	})

	return &AppsyncClient{
		transport: connection_output.Connection,
		runtime:   session,
	}, nil
}

type AppsyncClient struct {
	transport connection.Connection
	runtime   *engine.Engine
}

func (client *AppsyncClient) Publish(ctx context.Context, input PublishCommandInput) (*PublishCommandOutput, error) {
	return nil, nil
}

func (client *AppsyncClient) Subscribe(ctx context.Context, input SubscribeCommandInput) (*SubscribeCommandOutput, error) {
	return nil, nil
}

func (client *AppsyncClient) Close(ctx context.Context) error {
	client.runtime.Close(ctx)
	client.transport.Close()
	return nil
}
