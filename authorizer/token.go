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

func Token(auth_token string, endpoint *url.URL) port.Authorizer {
	return &token_authorizer{
		auth_token: auth_token,
		endpoint:   endpoint,
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
