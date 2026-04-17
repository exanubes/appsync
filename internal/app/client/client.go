package client

import (
	"context"
	"fmt"
	"time"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/engine"
	heartbeat "github.com/exanubes/appsync/internal/app/hearbeat"
	"github.com/exanubes/appsync/internal/app/protocol"
)

var connection_init_msg = []byte(`{"type":"connection_init"}`)

type ClientOptions struct{}

type AppsyncClient struct {
	connection app.Connection
	logger     app.Logger
	codec      app.Codec
	authorizer app.Authorizer
}

func New(
	connection app.Connection,
	codec app.Codec,
	authorizer app.Authorizer,
	logger app.Logger,
) *AppsyncClient {
	return &AppsyncClient{
		connection: connection,
		logger:     logger,
		codec:      codec,
		authorizer: authorizer,
	}
}

func (client *AppsyncClient) Connect(ctx context.Context) (*engine.Engine, error) {
	data, err := client.codec.Encode(connection_init_msg, nil)
	if err != nil {
		return nil, err
	}

	err = client.connection.Write(ctx, data)

	if err != nil {
		client.logger.Debug("Failed to send connection init message")
		return nil, err
	}
	client.logger.Debug("Connection init message sent successfully")

	timeout := time.After(10 * time.Second)
	messages_to_process := make([]app.Message, 0)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, app.ErrHandshakeTimeout
		default:
		}

		event, err := client.connection.Read(ctx)
		if err != nil {
			return nil, err
		}

		msg, err := client.codec.Decode(event)
		if err != nil {
			return nil, err
		}

		if msg, ok := msg.(protocol.ConnectionAckMessage); ok {
			client.logger.Debug("Connection Acknowledged")
			connection_timeout_ms := time.Duration(msg.TimeoutMs) * time.Millisecond

			return engine.New(heartbeat.New(connection_timeout_ms)), nil
		}

		if msg, ok := msg.(protocol.ErrorMessage); ok {
			return nil, fmt.Errorf("Handshake returned with error: %v", msg.Errors)
		}

		messages_to_process = append(messages_to_process, msg)
	}
}
