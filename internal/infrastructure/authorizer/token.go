package authorizer

import (
	"context"
	"net/url"

	"github.com/exanubes/appsync/internal/app"
)

type TokenAuthorizer struct {
	auth_token string
	endpoint   *url.URL
}

func NewTokenAuthorizer(auth_token string, endpoint *url.URL) *TokenAuthorizer {
	return &TokenAuthorizer{
		auth_token: auth_token,
		endpoint:   endpoint,
	}
}

func (authorizer *TokenAuthorizer) Authorize(ctx context.Context, input app.AuthorizeCommandInput) (app.Signature, error) {
	return app.Signature{
		"Authorization": authorizer.auth_token,
		"host":          authorizer.endpoint.Host,
	}, nil
}
