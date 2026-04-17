package connection

import (
	"context"
	"fmt"
	"time"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
)

var connection_init_msg = []byte(`{"type":"connection_init"}`)

type AuthorizeConnectionService struct {
	codec      app.Codec
	authorizer app.RequestAuthorizer
	logger     app.Logger
}

func NewAuthorizeConnectionService(codec app.Codec, authorizer app.RequestAuthorizer, logger app.Logger) *AuthorizeConnectionService {
	return &AuthorizeConnectionService{codec, authorizer, logger}
}

func (session *AuthorizeConnectionService) Authorize(ctx context.Context, connection Connection) (time.Duration, error) {
	data, err := session.codec.Encode(connection_init_msg, nil)
	if err != nil {
		return 0, err
	}

	session.logger.Debug("Initializing connection...")
	err = connection.Write(ctx, data)

	if err != nil {
		session.logger.Debug("Failed to send connection init message")
		return 0, err
	}

	timeout := time.After(10 * time.Second)
	messages_to_process := make([]app.Message, 0)
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-timeout:
			return 0, app.ErrHandshakeTimeout
		default:
		}

		event, err := connection.Read(ctx)
		if err != nil {
			return 0, err
		}

		msg, err := session.codec.Decode(event)
		if err != nil {
			return 0, err
		}

		if msg, ok := msg.(protocol.ConnectionAckMessage); ok {
			session.logger.Debug("Connection Acknowledged")
			connection_timeout_ms := time.Duration(msg.TimeoutMs) * time.Millisecond

			return connection_timeout_ms, nil
		}

		if msg, ok := msg.(protocol.ErrorMessage); ok {
			return 0, fmt.Errorf("Handshake returned with error: %v", msg.Errors)
		}

		messages_to_process = append(messages_to_process, msg)
	}
}
