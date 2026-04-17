package appsync

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/client"
	"github.com/exanubes/appsync/internal/infrastructure/authorizer"
	"github.com/exanubes/appsync/internal/infrastructure/codec"
	"github.com/exanubes/appsync/internal/infrastructure/logger"
	"github.com/exanubes/appsync/internal/infrastructure/transport"
)

// Connect establishes a new AppSync WebSocket connection and returns a Client.
func Connect(ctx context.Context) (*AppsyncClient, error) {
	websocket_connection, err := transport.Dial(ctx, app.DialOptions{})

	if err != nil {
		return nil, err
	}
	c := client.New(websocket_connection, codec.New(), authorizer.NewIAMAuthorizer(), logger.New())
	_, err := c.Connect(ctx)

	if err != nil {
		return nil, err
	}

	return &AppsyncClient{}, nil
}

type AppsyncClient struct {
	engine app.Engine
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

func (client *AppsyncClient) Err(ctx context.Context, input PublishCommandInput) error {
	return nil
}
