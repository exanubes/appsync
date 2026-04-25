package appsync

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/exanubes/appsync/internal/app"
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

type builder struct {
	errors       []error
	endpoint     *url.URL
	authorizer   app.RequestAuthorizer
	subprotocols []string
	logger       app.Logger
}

// Creates a ConnectionBuilder for a step by step configuration
func New() *builder {
	return &builder{}
}

// Parses endpoint into a URL
func (builder *builder) WithEndpoint(endpoint string) *builder {
	ws_endpoint, err := url.Parse(endpoint)

	if err != nil {
		builder.errors = append(builder.errors, err)
	}

	builder.endpoint = ws_endpoint

	return builder
}

// Sets authorizer
// Options: IAM, Lambda, Open ID Connect, Cognito User Pool, API Key
func (builder *builder) WithAuthorizer(authorizer Authorizer) *builder {
	if authorizer == nil {
		builder.errors = append(builder.errors, errors.New("Authorizer can't be nil"))
		return builder
	}
	request_authorizer, ok := authorizer.(*authorizer_impl)

	if !ok {
		builder.errors = append(builder.errors, errors.New("Invalid authorizer"))
	}

	builder.authorizer = request_authorizer.impl

	return builder
}

// Sets subprotocols
func (builder *builder) WithSubprotocol(subprotocols ...string) *builder {
	builder.subprotocols = subprotocols
	return builder
}

// Sets logger
func (builder *builder) WithLogger(logger app.Logger) *builder {
	builder.logger = logger
	return builder
}

// Validates inputs and creates a websocket connection
func (builder *builder) Connect(ctx context.Context) (*AppsyncClient, error) {
	if len(builder.errors) != 0 {
		return nil, fmt.Errorf("Invalid configuration, could not create connection: %+v", builder.errors)
	}

	if builder.authorizer == nil {
		return nil, errors.New("Authorizer is required")
	}

	if builder.logger == nil {
		builder.logger = logger.NoopLogger{}
	}

	dialer := transport.New()
	msg_codec := codec.New()
	base64_serializer := serializer.New()

	generate_subprotocol_service := connection.NewGenerateSubprotocolService(builder.authorizer, base64_serializer)
	authorize_connection_service := connection.NewAuthorizeConnectionService(msg_codec, builder.authorizer, builder.logger)
	create_connection_service := connection.NewConnectionService(dialer, authorize_connection_service, generate_subprotocol_service, builder.logger)

	connection_output, err := create_connection_service.Connect(ctx, connection.CreateConnectionInput{
		Url:          builder.endpoint,
		Subprotocols: builder.subprotocols,
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
		builder.authorizer,
		ingress_queue,
		egress_queue,
		pending_registry,
	)
	msg_router := router.New(pending_registry, usecases.ReceiveData)
	runtime := runtime.New(ingress_queue, msg_router, heartbeat)
	session := engine.New(heartbeat, runtime, io_loops, builder.logger)
	session.Start(ctx, engine.StartEngineInput{
		Timeout: connection_output.Timeout,
	})

	return &AppsyncClient{
		transport: connection_output.Connection,
		runtime:   session,
		usecases:  usecases,
	}, nil
}

// Connect establishes a new AppSync WebSocket connection and returns a Client.
func Connect(ctx context.Context, options ConnectionOptions) (*AppsyncClient, error) {
	builder := New()

	builder.
		WithAuthorizer(options.Authorizer).
		WithEndpoint(options.Endpoint).
		WithLogger(options.Logger).
		WithSubprotocol(options.Subprotocols...)

	return builder.Connect(ctx)
}
