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
	codec      app.Decoder
	authorizer app.RequestAuthorizer
	logger     app.Logger
}

func NewAuthorizeConnectionService(codec app.Decoder, authorizer app.RequestAuthorizer, logger app.Logger) *AuthorizeConnectionService {
	return &AuthorizeConnectionService{
		codec:      codec,
		authorizer: authorizer,
		logger:     logger.SetContext("AuthorizeConnectionService")}
}

func (session *AuthorizeConnectionService) Authorize(ctx context.Context, connection Connection) (time.Duration, error) {
	session.logger.Debug("Initializing connection...")
	err := connection.Write(ctx, connection_init_msg)

	if err != nil {
		session.logger.Debug("Failed to send connection init message")
		return 0, err
	}

	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-timeout:
			return 0, app.ErrHandshakeTimeout
		default:
		}

		session.logger.Debug("Waiting for message...")
		event, err := connection.Read(ctx)
		if err != nil {
			session.logger.Debug("Fail")
			return 0, err
		}
		session.logger.Debug("Success", "data", string(event))

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

		// Skip
	}
}
