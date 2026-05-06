package authorizer

import (
	"context"
	"net/url"

	"github.com/exanubes/appsync/port"
)

type api_key_authorizer struct {
	api_key  string
	endpoint *url.URL
}

type ApiKeyAuthorizerConfig struct {
	ApiKey   string
	Endpoint *url.URL
}

func ApiKey(config ApiKeyAuthorizerConfig) port.Authorizer {
	return &api_key_authorizer{endpoint: config.Endpoint, api_key: config.ApiKey}

}

func (authorizer *api_key_authorizer) Authorize(ctx context.Context, input port.AuthorizeCommandInput) (*port.AuthorizeCommandOutput, error) {
	return &port.AuthorizeCommandOutput{
		Signature: map[string]string{
			"host":      authorizer.endpoint.Host,
			"x-api-key": authorizer.api_key,
		},
	}, nil
}
