package appsync

import (
	"context"
	"net/url"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/infrastructure/authorizer"
	"github.com/exanubes/appsync/internal/infrastructure/clock"
)

type authorizer_adapter struct {
	impl app.RequestAuthorizer
}

func (authorizer *authorizer_adapter) Authorize(ctx context.Context, input AuthorizeCommandInput) (Signature, error) {
	result, err := authorizer.impl.Authorize(ctx, app.AuthorizeCommandInput{
		Channel: input.Channel,
		Payload: input.Payload,
	})

	if err != nil {
		return nil, err
	}

	return Signature(result), nil
}

func NewIAMAuthorizer(region string, endpoint *url.URL) Authorizer {
	credentials_provider := authorizer.NewAwsCredentialsProvider()
	clock := clock.New()
	signer := authorizer.NewSigv4Signer(region, credentials_provider, clock)

	return &authorizer_adapter{
		impl: authorizer.NewIAMAuthorizer(
			endpoint,
			signer,
			authorizer.NewCanonicalRequestFactory(),
		)}
}

func NewApiKeyAuthorizer(api_key string, endpoint *url.URL) Authorizer {
	return &authorizer_adapter{
		impl: authorizer.NewApiKeyAuthorizer(
			api_key,
			endpoint,
		)}
}

func NewLambdaAuthorizer(token string, endpoint *url.URL) Authorizer {
	return &authorizer_adapter{
		impl: authorizer.NewLambdaAuthorizer(
			token,
			endpoint,
		)}
}
