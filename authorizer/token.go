package authorizer

import (
	"context"
	"net/url"

	"github.com/exanubes/appsync/port"
)

type token_authorizer struct {
	auth_token string
	endpoint   *url.URL
}

type TokenAuthorizerConfig struct {
	AuthToken string
	Endpoint  *url.URL
}

func Token(config TokenAuthorizerConfig) port.Authorizer {
	return &token_authorizer{
		auth_token: config.AuthToken,
		endpoint:   config.Endpoint,
	}
}

func (authorizer *token_authorizer) Authorize(ctx context.Context, input port.AuthorizeCommandInput) (*port.AuthorizeCommandOutput, error) {
	return &port.AuthorizeCommandOutput{
		Signature: map[string]string{
			"Authorization": authorizer.auth_token,
			"host":          authorizer.endpoint.Host,
		},
	}, nil
}
