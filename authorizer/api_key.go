package authorizer

import (
	"context"
	"net/url"
)

type api_key_authorizer struct {
	api_key  string
	endpoint *url.URL
}

type ApiKeyAuthorizerConfig struct {
	ApiKey   string
	Endpoint string
}

func ApiKey(config ApiKeyAuthorizerConfig) (Authorizer, error) {
	endpoint, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	return &api_key_authorizer{endpoint: endpoint, api_key: config.ApiKey}, nil
}

func (authorizer *api_key_authorizer) Authorize(ctx context.Context, input AuthorizeCommandInput) (*AuthorizeCommandOutput, error) {
	return &AuthorizeCommandOutput{
		Signature: map[string]string{
			"host":      authorizer.endpoint.Host,
			"x-api-key": authorizer.api_key,
		},
	}, nil
}
