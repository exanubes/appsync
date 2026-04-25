package authorizer

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type LambdaAuthorizer struct{}

func NewLambdaAuthorizer() *LambdaAuthorizer {
	return &LambdaAuthorizer{}
}

func (authorizer *LambdaAuthorizer) Authorize(ctx context.Context, input app.AuthorizeCommandInput) (app.Signature, error) {
	return nil, nil
}
