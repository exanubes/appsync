package appsync

import (
	"context"
	"errors"
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
	"github.com/exanubes/appsync/authorizer"
	infra_authorizer "github.com/exanubes/appsync/internal/infrastructure/authorizer"
	"github.com/exanubes/appsync/internal/infrastructure/clock"
	"github.com/exanubes/appsync/internal/infrastructure/codec"
	"github.com/exanubes/appsync/internal/infrastructure/logger"
	"github.com/exanubes/appsync/internal/infrastructure/serializer"
	"github.com/exanubes/appsync/internal/infrastructure/transport"
)

type builder struct {
	errors       []error
	endpoint     *url.URL
	authorizer   authorizer.Authorizer
	subprotocols []string
	logger       app.Logger
	backpressure Backpressure
}

var default_backpressure_config = Backpressure{
	ConnectionInbound:  100,
	ConnectionOutbound: 100,
	SubscriptionEvents: 100,
}

// Creates a ConnectionBuilder for a step by step configuration
func new_builder() *builder {
	return &builder{
		logger:       logger.NoopLogger{},
		backpressure: default_backpressure_config,
	}
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
func (builder *builder) WithAuthorizer(authz authorizer.Authorizer) *builder {
	if authz == nil {
		builder.errors = append(builder.errors, errors.New("Authorizer can't be nil"))
		return builder
	}

	builder.authorizer = authz

	return builder
}

// Sets subprotocols
func (builder *builder) WithSubprotocol(subprotocols ...string) *builder {
	builder.subprotocols = subprotocols
	return builder
}

// Sets logger
func (builder *builder) WithLogger(logger app.Logger) *builder {
	if logger != nil {
		builder.logger = logger
	}
	return builder
}

func (builder *builder) WithBackpressure(config Backpressure) *builder {
	if config.ConnectionInbound != 0 {
		builder.backpressure.ConnectionInbound = config.ConnectionInbound

	}

	if config.ConnectionOutbound != 0 {
		builder.backpressure.ConnectionOutbound = config.ConnectionOutbound

	}

	if config.SubscriptionEvents != 0 {
		builder.backpressure.SubscriptionEvents = config.SubscriptionEvents

	}
	return builder
}

// Validates inputs and creates a websocket connection
func (builder *builder) Connect(ctx context.Context) (*appsync_client, error) {
	if len(builder.errors) != 0 {
		return nil, errors.Join(builder.errors...)
	}

	if builder.authorizer == nil {
		return nil, errors.New("Authorizer is required")
	}

	dialer := transport.New()
	msg_codec := codec.New()
	base64_serializer := serializer.New()
	request_authorizer := infra_authorizer.NewInternalAdapter(builder.authorizer)
	generate_subprotocol_service := connection.NewGenerateSubprotocolService(request_authorizer, base64_serializer)
	authorize_connection_service := connection.NewAuthorizeConnectionService(msg_codec, request_authorizer, builder.logger)
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
	ingress_queue := queue.NewIngressQueue(builder.backpressure.ConnectionInbound)
	egress_queue := queue.NewEgressQueue(builder.backpressure.ConnectionOutbound)
	pending_registry := pending.NewRegistry()
	io_loops := io.New(ingress_queue, egress_queue, connection_output.Connection, msg_codec)
	usecases := composition.NewUseCases(
		request_authorizer,
		ingress_queue,
		egress_queue,
		pending_registry,
		builder.backpressure.SubscriptionEvents,
	)
	msg_router := router.New(pending_registry, usecases.ReceiveData)
	runtime := runtime.New(ingress_queue, msg_router, heartbeat)
	session := engine.New(heartbeat, runtime, io_loops, builder.logger)
	session.Start(ctx, engine.StartEngineInput{
		Timeout: connection_output.Timeout,
	})

	return &appsync_client{
		transport: connection_output.Connection,
		runtime:   session,
		usecases:  usecases,
	}, nil
}

// Connect establishes a new AppSync WebSocket connection and returns a Client.
func Connect(ctx context.Context, options ConnectionOptions) (Client, error) {
	builder := new_builder()

	builder.
		WithAuthorizer(options.Authorizer).
		WithEndpoint(options.Endpoint).
		WithSubprotocol(options.Subprotocols...).
		WithBackpressure(options.Backpressure)

	return builder.Connect(ctx)
}
