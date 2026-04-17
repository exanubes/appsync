package authorizer

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type IAMAuthorizer struct{}

func NewIAMAuthorizer() *IAMAuthorizer {
	return &IAMAuthorizer{}
}
func (authorizer *IAMAuthorizer) Authorize(ctx context.Context, payload app.Payload) (app.Signature, error) {
	return nil, nil
}
