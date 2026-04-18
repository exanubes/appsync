package composition

import (
	"net/url"

	"github.com/exanubes/appsync/internal/infrastructure/authorizer"
)

func NewIAMAuthorizer(region string, endpoint *url.URL) *authorizer.IAMAuthorizer {
	credentials_provider := authorizer.NewAwsCredentialsProvider()
	signer := authorizer.NewSigv4Signer(region, credentials_provider)

	return authorizer.NewIAMAuthorizer(
		endpoint,
		signer,
		authorizer.NewCanonicalRequestFactory(),
	)
}
