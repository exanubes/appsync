package authorizer

import (
	"context"
	"net/url"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/infrastructure/authorizer/internal"
)

type IAMAuthorizer struct {
	signer  internal.Signer
	request internal.RequestFactory

	endpoint *url.URL
}

func NewIAMAuthorizer(endpoint *url.URL, signer internal.Signer, factory internal.RequestFactory) *IAMAuthorizer {
	return &IAMAuthorizer{
		signer:   signer,
		request:  factory,
		endpoint: endpoint,
	}
}
func (authorizer *IAMAuthorizer) Authorize(ctx context.Context, payload app.Payload) (app.Signature, error) {
	req, err := authorizer.request.Create(authorizer.endpoint, payload)

	if err != nil {
		return nil, err
	}

	signature, err := authorizer.signer.Sign(ctx, req)

	if err != nil {
		return nil, err
	}

	return signature, nil
}
