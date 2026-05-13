package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
)

func main() {
	lambda.Start(func(ctx context.Context, event events.AppSyncLambdaAuthorizerRequest) (events.AppSyncLambdaAuthorizerResponse, error) {
		expectedToken := os.Getenv("LAMBDA_AUTHORIZER_TOKEN")

		if expectedToken == "" || event.AuthorizationToken != expectedToken {
			return events.AppSyncLambdaAuthorizerResponse{
				IsAuthorized: false,
				TTLOverride:  aws.Int(0),
			}, nil
		}

		return events.AppSyncLambdaAuthorizerResponse{
			IsAuthorized: true,
			ResolverContext: map[string]interface{}{
				"userId":     "1234",
				"authorizer": "lambda",
			},
			TTLOverride: aws.Int(300),
		}, nil
	})
}
