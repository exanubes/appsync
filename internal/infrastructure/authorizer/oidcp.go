package authorizer

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type OidcpAuthorizer struct{}

func NewOidcpAuthorizer() *OidcpAuthorizer {
	return &OidcpAuthorizer{}
}

func (authorizer *OidcpAuthorizer) Authorize(ctx context.Context, input app.AuthorizeCommandInput) (app.Signature, error) {
	return nil, nil
}
