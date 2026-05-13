package authorizer_test

import (
	"context"
	"log"

	"github.com/exanubes/appsync"
	"github.com/exanubes/appsync/authorizer"
)

func ExampleApiKey() {
	ctx := context.Background()

	authz, err := authorizer.ApiKey(authorizer.ApiKeyAuthorizerConfig{
		ApiKey:   "da2-xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		Endpoint: "https://xxxxxxxxxxxxxxxxxxxx.appsync-api.us-east-1.amazonaws.com",
	})
	if err != nil {
		log.Fatal(err)
	}

	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     "wss://xxxxxxxxxxxxxxxxxxxx.appsync-realtime-api.us-east-1.amazonaws.com",
		Subprotocols: []string{appsync.ProtocolEvents},
		Authorizer:   authz,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close(ctx)
}

func ExampleIAM() {
	ctx := context.Background()

	authz, err := authorizer.IAM(authorizer.IAMAuthorizerConfig{
		Region:   "us-east-1",
		Endpoint: "https://xxxxxxxxxxxxxxxxxxxx.appsync-api.us-east-1.amazonaws.com",
	})
	if err != nil {
		log.Fatal(err)
	}

	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     "wss://xxxxxxxxxxxxxxxxxxxx.appsync-realtime-api.us-east-1.amazonaws.com",
		Subprotocols: []string{appsync.ProtocolEvents},
		Authorizer:   authz,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close(ctx)
}

func ExampleToken() {
	ctx := context.Background()

	// Works with Cognito ID tokens, OIDC tokens, and Lambda authorizer tokens.
	authz, err := authorizer.Token(authorizer.TokenAuthorizerConfig{
		AuthToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
		Endpoint:  "https://xxxxxxxxxxxxxxxxxxxx.appsync-api.us-east-1.amazonaws.com",
	})
	if err != nil {
		log.Fatal(err)
	}

	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     "wss://xxxxxxxxxxxxxxxxxxxx.appsync-realtime-api.us-east-1.amazonaws.com",
		Subprotocols: []string{appsync.ProtocolEvents},
		Authorizer:   authz,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close(ctx)
}
