package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
)

func main() {
	lambda.Start(func(ctx context.Context, event events.AppSyncLambdaAuthorizerRequest) (events.AppSyncLambdaAuthorizerResponse, error) {
		token := event.AuthorizationToken
		request := event.RequestContext
		fmt.Println("@@@")
		fmt.Printf("TOKEN: %+v\n", token)
		fmt.Println("----")
		fmt.Printf("REQUEST: %+v\n", request)
		fmt.Println("----")
		fmt.Printf("EVENT: %+v\n", event)
		fmt.Println("###")

		authorized := true

		if token != "custom-token" {
			authorized = false
		}

		return events.AppSyncLambdaAuthorizerResponse{
			IsAuthorized: authorized,
			ResolverContext: map[string]interface{}{
				"userId": "1234",
			},
			TTLOverride: aws.Int(300),
		}, nil
	})
}
