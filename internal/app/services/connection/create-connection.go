package connection

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type ConnectionService struct {
	dialer      Dialer
	authorizer  ConnectionAuthorizer
	subprotocol SubprotocolGenerator
	logger      app.Logger
}

func NewConnectionService(dialer Dialer, connection_authorizer ConnectionAuthorizer, subprotocol_generator SubprotocolGenerator, logger app.Logger) *ConnectionService {
	return &ConnectionService{
		dialer:      dialer,
		authorizer:  connection_authorizer,
		subprotocol: subprotocol_generator,
		logger:      logger.SetContext("CreateConnectionService"),
	}
}

func (service *ConnectionService) Connect(ctx context.Context, input CreateConnectionInput) (*CreateConnectionOutput, error) {
	subprotocol, err := service.subprotocol.Generate(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := service.dialer.Dial(ctx, DialOptions{
		Url:          input.Url,
		Subprotocols: append(input.Subprotocols, subprotocol),
	})

	if err != nil {
		return nil, err
	}

	timeout, err := service.authorizer.Authorize(ctx, conn)

	if err != nil {
		return nil, err
	}

	return &CreateConnectionOutput{
		Connection: conn,
		Timeout:    timeout,
	}, nil
}
