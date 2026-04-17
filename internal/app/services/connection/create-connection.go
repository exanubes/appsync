package connection

import "context"

type ConnectionService struct {
	dialer      Dialer
	authorizer  ConnectionAuthorizer
	subprotocol SubprotocolGenerator
}

func NewConnectionService(dialer Dialer, connection_authorizer ConnectionAuthorizer, subprotocol_generator SubprotocolGenerator) *ConnectionService {
	return &ConnectionService{
		dialer:     dialer,
		authorizer: connection_authorizer,
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
