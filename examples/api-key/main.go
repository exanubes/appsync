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

func main() {
	var (
		httpEndpoint = requiredEnv("APPSYNC_HTTP_ENDPOINT")
		wsEndpoint   = requiredEnv("APPSYNC_WS_ENDPOINT")
		apiKey       = requiredEnv("APPSYNC_API_KEY")
		channel      = requiredEnv("APPSYNC_CHANNEL")
	)

	ctx := context.Background()

	authz, err := authorizer.ApiKey(authorizer.ApiKeyAuthorizerConfig{
		ApiKey:   apiKey,
		Endpoint: httpEndpoint,
	})
	if err != nil {
		log.Fatal(err)
	}

	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     wsEndpoint,
		Subprotocols: []string{appsync.ProtocolEvents},
		Authorizer:   authz,
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

	payload, err := json.Marshal(event{Message: "hello from api-key example"})
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

	fmt.Println("received:", received.Message)
}
