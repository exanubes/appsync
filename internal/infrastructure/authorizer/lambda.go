package authorizer

import (
	"context"
	"net/url"

	"github.com/exanubes/appsync/internal/app"
)

type LambdaAuthorizer struct {
	token    string
	endpoint *url.URL
}

func NewLambdaAuthorizer(token string, endpoint *url.URL) *LambdaAuthorizer {
	return &LambdaAuthorizer{
		token:    token,
		endpoint: endpoint,
	}
}

func (authorizer *LambdaAuthorizer) Authorize(ctx context.Context, input app.AuthorizeCommandInput) (app.Signature, error) {
	return app.Signature{"Authorization": authorizer.token,
		"host": authorizer.endpoint.Host}, nil
}
