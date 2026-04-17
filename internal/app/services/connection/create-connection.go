package connection

import "context"

type ConnectionService struct {
	dialer     Dialer
	authorizer ConnectionAuthorizer
}

func NewConnectionService(dialer Dialer, connection_authorizer ConnectionAuthorizer) *ConnectionService {
	return &ConnectionService{
		dialer:     dialer,
		authorizer: connection_authorizer,
	}
}

func (service *ConnectionService) Connect(ctx context.Context, input CreateConnectionInput) (*CreateConnectionOutput, error) {
	conn, err := service.dialer.Dial(ctx, DialOptions{
		Url:          input.Url,
		Subprotocols: input.Subprotocols,
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
