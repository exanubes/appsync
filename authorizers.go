package appsync

import (
	"context"
	"net/url"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/infrastructure/authorizer"
)

type authorizer_impl struct {
	impl app.RequestAuthorizer
}

func (authorizer *authorizer_impl) Authorize(ctx context.Context, input AuthorizeCommandInput) (Signature, error) {
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
	signer := authorizer.NewSigv4Signer(region, credentials_provider)

	return &authorizer_impl{
		impl: authorizer.NewIAMAuthorizer(
			endpoint,
			signer,
			authorizer.NewCanonicalRequestFactory(),
		)}
}
