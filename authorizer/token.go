package authorizer

import (
	"context"
	"net/url"
)

type token_authorizer struct {
	auth_token string
	endpoint   *url.URL
}

type TokenAuthorizerConfig struct {
	AuthToken string
	Endpoint  string
}

func Token(config TokenAuthorizerConfig) (Authorizer, error) {
	endpoint, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	return &token_authorizer{
		auth_token: config.AuthToken,
		endpoint:   endpoint,
	}, nil
}

func (authorizer *token_authorizer) Authorize(ctx context.Context, input AuthorizeCommandInput) (*AuthorizeCommandOutput, error) {
	return &AuthorizeCommandOutput{
		Signature: map[string]string{
			"Authorization": authorizer.auth_token,
			"host":          authorizer.endpoint.Host,
		},
	}, nil
}
