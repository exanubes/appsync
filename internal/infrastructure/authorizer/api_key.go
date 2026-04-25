package authorizer

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type ApiKeyAuthorizer struct{}

func NewApiAuthorizer() *ApiKeyAuthorizer {
	return &ApiKeyAuthorizer{}
}

func (authorizer *ApiKeyAuthorizer) Authorize(ctx context.Context, input app.AuthorizeCommandInput) (app.Signature, error) {
	return nil, nil
}
