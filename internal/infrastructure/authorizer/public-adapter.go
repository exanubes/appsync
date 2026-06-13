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
	payload := make([][]byte, len(input.Payload))
	for index := range input.Payload {
		payload[index] = []byte(input.Payload[index])
	}

	output, err := authorizer.public.Authorize(ctx, pub.AuthorizeCommandInput{
		Channel: input.Channel,
		Payload: payload,
	})

	if err != nil {
		return nil, err
	}

	return output.Signature, nil
}
