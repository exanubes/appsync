package authorizer

import "github.com/exanubes/appsync/internal/infrastructure/authorizer/internal"

func NewCanonicalRequestFactory() internal.CanonicalRequest {
	return internal.CanonicalRequest{}
}
