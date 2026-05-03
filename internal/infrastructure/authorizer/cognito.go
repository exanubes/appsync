package authorizer

import (
	"context"
	"net/url"

	"github.com/exanubes/appsync/internal/app"
)

type CognitoAuthorizer struct {
	auth_token string
	endpoint   *url.URL
}

func NewCognitoAuthorizer(auth_token string, endpoint *url.URL) *CognitoAuthorizer {
	return &CognitoAuthorizer{
		auth_token: auth_token,
		endpoint:   endpoint,
	}
}

func (authorizer *CognitoAuthorizer) Authorize(ctx context.Context, input app.AuthorizeCommandInput) (app.Signature, error) {
	return app.Signature{
		"Authorization": authorizer.auth_token,
		"host":          authorizer.endpoint.Host,
	}, nil
}
