package appsync

import (
	"context"
	"net/url"

	"github.com/exanubes/appsync/internal/app/engine"
	"github.com/exanubes/appsync/internal/app/heartbeat"
	"github.com/exanubes/appsync/internal/app/pending"
	"github.com/exanubes/appsync/internal/app/queue"
	"github.com/exanubes/appsync/internal/app/router"
	"github.com/exanubes/appsync/internal/app/runtime"
	"github.com/exanubes/appsync/internal/app/services/connection"
	"github.com/exanubes/appsync/internal/app/services/io"
	"github.com/exanubes/appsync/internal/composition"
	"github.com/exanubes/appsync/internal/infrastructure/clock"
	"github.com/exanubes/appsync/internal/infrastructure/codec"
	"github.com/exanubes/appsync/internal/infrastructure/logger"
	"github.com/exanubes/appsync/internal/infrastructure/serializer"
	"github.com/exanubes/appsync/internal/infrastructure/transport"
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

	clock := clock.New()
	heartbeat := heartbeat.New(clock)
	ingress_queue := queue.NewIngressQueue(100)
	egress_queue := queue.NewEgressQueue(100)
	pending_registry := pending.NewRegistry()
	io_loops := io.New(ingress_queue, egress_queue, connection_output.Connection, msg_codec)
	usecases := composition.NewUseCases(
		request_authorizer,
		ingress_queue,
		egress_queue,
		pending_registry,
	)
	msg_router := router.New(pending_registry, usecases.ReceiveData)
	runtime := runtime.New(ingress_queue, msg_router, heartbeat)
	session := engine.New(heartbeat, runtime, io_loops, slogger)
	session.Start(ctx, engine.StartEngineInput{
		Timeout: connection_output.Timeout,
	})

	return &AppsyncClient{
		transport: connection_output.Connection,
		runtime:   session,
		usecases:  usecases,
	}, nil
}
