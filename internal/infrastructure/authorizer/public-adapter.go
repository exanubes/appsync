package authorizer

import (
	"context"

	pub "github.com/exanubes/appsync/authorizer"
	"github.com/exanubes/appsync/internal/app"
)

type InternalAuthorizerAdapter struct {
	public pub.Authorizer
}

func NewInternalAdapter(public pub.Authorizer) *InternalAuthorizerAdapter {
	return &InternalAuthorizerAdapter{
		public: public,
	}
}

func (authorizer *InternalAuthorizerAdapter) Authorize(ctx context.Context, input app.AuthorizeCommandInput) (app.Signature, error) {
	output, err := authorizer.public.Authorize(ctx, pub.AuthorizeCommandInput{
		Channel: input.Channel,
		Payload: input.Payload,
	})

	if err != nil {
		return nil, err
	}

	return output.Signature, nil
}
