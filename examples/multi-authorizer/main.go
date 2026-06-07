package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/exanubes/appsync"
	"github.com/exanubes/appsync/authorizer"
)

func requiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}

// This example demonstrates using different authorizers for different operations.
//
// Scenario: a channel that accepts API key authorization for subscribing but requires
// a Lambda token to publish. A third per-request token is used for a single publish call.
//
// Environment variables:
//   APPSYNC_HTTP_ENDPOINT   - AppSync HTTP endpoint
//   APPSYNC_WS_ENDPOINT     - AppSync WebSocket realtime endpoint
//   APPSYNC_API_KEY         - API key (used for subscribe)
//   APPSYNC_PUBLISH_TOKEN   - Lambda/Cognito token (used for publish)
//   APPSYNC_OVERRIDE_TOKEN  - Token used to override authorizer on a single publish
//   APPSYNC_CHANNEL         - Channel to subscribe and publish to

func main() {
	var (
		httpEndpoint  = requiredEnv("APPSYNC_HTTP_ENDPOINT")
		wsEndpoint    = requiredEnv("APPSYNC_WS_ENDPOINT")
		apiKey        = requiredEnv("APPSYNC_API_KEY")
		publishToken  = requiredEnv("APPSYNC_PUBLISH_TOKEN")
		overrideToken = requiredEnv("APPSYNC_OVERRIDE_TOKEN")
		channel       = requiredEnv("APPSYNC_CHANNEL")
	)

	ctx := context.Background()

	// connectAuthz and subscribeAuthz use API key authorization.
	connectAuthz, err := authorizer.ApiKey(authorizer.ApiKeyAuthorizerConfig{
		ApiKey:   apiKey,
		Endpoint: httpEndpoint,
	})
	if err != nil {
		log.Fatal(err)
	}

	// publishAuthz uses a token (e.g. Lambda authorizer) for publish operations.
	publishAuthz, err := authorizer.Token(authorizer.TokenAuthorizerConfig{
		AuthToken: publishToken,
		Endpoint:  httpEndpoint,
	})
	if err != nil {
		log.Fatal(err)
	}

	// overrideAuthz is a one-off authorizer passed directly to a single Publish call.
	overrideAuthz, err := authorizer.Token(authorizer.TokenAuthorizerConfig{
		AuthToken: overrideToken,
		Endpoint:  httpEndpoint,
	})
	if err != nil {
		log.Fatal(err)
	}

	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     wsEndpoint,
		Subprotocols: []string{appsync.ProtocolEvents},
		Authorizers: appsync.Authorizers{
			Connect:   connectAuthz,
			Subscribe: connectAuthz,
			Publish:   publishAuthz,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close(ctx)

	sub, err := client.Subscribe(ctx, appsync.SubscribeCommandInput{
		Channel: channel,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Close(ctx)

	type event struct {
		Message string `json:"message"`
	}

	// Regular publish — uses Authorizers.Publish (publishAuthz).
	payload, err := json.Marshal(event{Message: "hello from publish authorizer"})
	if err != nil {
		log.Fatal(err)
	}

	if err = client.Publish(ctx, appsync.PublishCommandInput{
		Channel: channel,
		Payload: payload,
	}); err != nil {
		log.Fatal(err)
	}

	var received event
	if err = sub.DecodeNext(ctx, &received); err != nil {
		if errors.Is(err, appsync.ErrSubscriptionClosed) {
			return
		}
		log.Fatal(err)
	}
	fmt.Println("received (publish authorizer):", received.Message)

	// Per-request override — uses overrideAuthz for this single call only.
	overridePayload, err := json.Marshal(event{Message: "hello from override authorizer"})
	if err != nil {
		log.Fatal(err)
	}

	if err = client.Publish(ctx, appsync.PublishCommandInput{
		Channel:    channel,
		Payload:    overridePayload,
		Authorizer: overrideAuthz,
	}); err != nil {
		log.Fatal(err)
	}

	if err = sub.DecodeNext(ctx, &received); err != nil {
		if errors.Is(err, appsync.ErrSubscriptionClosed) {
			return
		}
		log.Fatal(err)
	}
	fmt.Println("received (override authorizer):", received.Message)
}
