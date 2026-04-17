package appsync

import (
	"context"

	"github.com/exanubes/appsync/internal/app/engine"
	"github.com/exanubes/appsync/internal/app/services/connection"
	"github.com/exanubes/appsync/internal/infrastructure/authorizer"
	"github.com/exanubes/appsync/internal/infrastructure/codec"
	"github.com/exanubes/appsync/internal/infrastructure/logger"
	"github.com/exanubes/appsync/internal/infrastructure/serializer"
	"github.com/exanubes/appsync/internal/infrastructure/transport"
)

// Connect establishes a new AppSync WebSocket connection and returns a Client.
func Connect(ctx context.Context, options ConnectionOptions) (*AppsyncClient, error) {
	slogger := logger.New()
	dialer := transport.New()
	request_authorizer := authorizer.NewIAMAuthorizer()
	msg_codec := codec.New()
	base64_serializer := serializer.New()

	generate_subprotocol_service := connection.NewGenerateSubprotocolService(request_authorizer, base64_serializer)
	authorize_connection_service := connection.NewAuthorizeConnectionService(msg_codec, request_authorizer, slogger)
	create_connection_service := connection.NewConnectionService(dialer, authorize_connection_service, generate_subprotocol_service)

	connection_output, err := create_connection_service.Connect(ctx, connection.CreateConnectionInput{
		Url:          options.Url,
		Subprotocols: options.Subprotocols,
	})

	if err != nil {
		return nil, err
	}

	runtime := engine.New()
	runtime.Start(ctx, engine.StartEngineInput{
		Timeout:    connection_output.Timeout,
		Connection: connection_output.Connection,
	})

	return &AppsyncClient{}, nil
}

type AppsyncClient struct {
}

func (client *AppsyncClient) Publish(ctx context.Context, input PublishCommandInput) (*PublishCommandOutput, error) {
	return nil, nil
}

func (client *AppsyncClient) Subscribe(ctx context.Context, input SubscribeCommandInput) (*SubscribeCommandOutput, error) {
	return nil, nil
}

func (client *AppsyncClient) Close(ctx context.Context) error {
	return nil
}
