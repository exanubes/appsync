package authorizer

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type CognitoAuthorizer struct{}

func NewCognitoAuthorizer() *CognitoAuthorizer {
	return &CognitoAuthorizer{}
}

func (authorizer *CognitoAuthorizer) Authorize(ctx context.Context, input app.AuthorizeCommandInput) (app.Signature, error) {
	return nil, nil
}
