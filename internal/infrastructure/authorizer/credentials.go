package authorizer

import "github.com/exanubes/appsync/internal/infrastructure/authorizer/internal"

func NewAwsCredentialsProvider() *internal.AwsCredentialsProvider {
	return &internal.AwsCredentialsProvider{}
}
