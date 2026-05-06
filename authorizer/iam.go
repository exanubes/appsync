package authorizer

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/exanubes/appsync/authorizer/internal"
	"github.com/exanubes/appsync/internal/infrastructure/clock"
	"github.com/exanubes/appsync/port"
)

type iam_authorizer struct {
	signer  internal.Signer
	request internal.RequestFactory

	endpoint *url.URL
}

type IAMAuthorizerConfig struct {
	Region   string
	Endpoint *url.URL
}

func IAM(config IAMAuthorizerConfig) port.Authorizer {
	credentials_provider := &internal.AwsCredentialsProvider{}
	clock := clock.New()
	signer := &internal.Sigv4Signer{Provider: credentials_provider, Region: config.Region, Clock: clock}
	return new_iam_authorizer(config.Endpoint, signer, internal.CanonicalRequest{})
}

func new_iam_authorizer(endpoint *url.URL, signer internal.Signer, factory internal.RequestFactory) *iam_authorizer {
	return &iam_authorizer{
		signer:   signer,
		request:  factory,
		endpoint: endpoint,
	}
}
func (authorizer *iam_authorizer) Authorize(ctx context.Context, input port.AuthorizeCommandInput) (*port.AuthorizeCommandOutput, error) {
	canonical := internal.CanonicalPayload{
		Channel: input.Channel,
	}

	if input.Payload != nil {
		canonical.Payload = []string{string(input.Payload)}
	}

	payload, err := json.Marshal(canonical)

	if err != nil {
		return nil, err
	}

	req, err := authorizer.request.Create(authorizer.endpoint, payload)

	if err != nil {
		return nil, err
	}

	signature, err := authorizer.signer.Sign(ctx, req)

	if err != nil {
		return nil, err
	}

	return &port.AuthorizeCommandOutput{
		Signature: signature,
	}, nil
}
