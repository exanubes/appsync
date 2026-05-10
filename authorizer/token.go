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
	Endpoint  string
}

func Token(config TokenAuthorizerConfig) (port.Authorizer, error) {
	endpoint, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	return &token_authorizer{
		auth_token: config.AuthToken,
		endpoint:   endpoint,
	}, nil
}

func (authorizer *token_authorizer) Authorize(ctx context.Context, input port.AuthorizeCommandInput) (*port.AuthorizeCommandOutput, error) {
	return &port.AuthorizeCommandOutput{
		Signature: map[string]string{
			"Authorization": authorizer.auth_token,
			"host":          authorizer.endpoint.Host,
		},
	}, nil
}
