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

func ApiKey(api_key string, endpoint *url.URL) port.Authorizer {
	return &api_key_authorizer{api_key, endpoint}

}

func (authorizer *api_key_authorizer) Authorize(ctx context.Context, input port.AuthorizeCommandInput) (*port.AuthorizeCommandOutput, error) {
	headers := map[string]string{
		"host":      authorizer.endpoint.Host,
		"x-api-key": authorizer.api_key,
	}

	return &port.AuthorizeCommandOutput{
		Signature: headers,
	}, nil
}
