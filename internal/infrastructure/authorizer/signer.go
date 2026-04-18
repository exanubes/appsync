package authorizer

import "github.com/exanubes/appsync/internal/infrastructure/authorizer/internal"

func NewSigv4Signer(region string, provider internal.CredentialProvider) *internal.Sigv4Signer {
	return &internal.Sigv4Signer{Provider: provider, Region: region}
}
