package authorizer

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/port"
)

type InternalAuthorizerAdapter struct {
	public port.Authorizer
}

func NewInternalAdapter(public port.Authorizer) *InternalAuthorizerAdapter {
	return &InternalAuthorizerAdapter{
		public: public,
	}
}

func (authorizer *InternalAuthorizerAdapter) Authorize(ctx context.Context, input app.AuthorizeCommandInput) (app.Signature, error) {
	output, err := authorizer.public.Authorize(ctx, port.AuthorizeCommandInput{
		Channel: input.Channel,
		Payload: input.Payload,
	})

	if err != nil {
		return nil, err
	}

	return output.Signature, nil
}
