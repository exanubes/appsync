package authorizer

import (
	"context"
	"net/url"

	"github.com/exanubes/appsync/internal/app"
)

type ApiKeyAuthorizer struct {
	api_key  string
	endpoint *url.URL
}

func NewApiKeyAuthorizer(api_key string, endpoint *url.URL) *ApiKeyAuthorizer {
	return &ApiKeyAuthorizer{api_key, endpoint}

}

func (authorizer *ApiKeyAuthorizer) Authorize(ctx context.Context, input app.AuthorizeCommandInput) (app.Signature, error) {
	headers := map[string]string{
		"host":      authorizer.endpoint.Host,
		"x-api-key": authorizer.api_key,
	}
	return headers, nil
}
