package appsync_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/exanubes/appsync"
	"github.com/exanubes/appsync/authorizer"
)

func ExampleConnect() {
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

func ExampleClient_Subscribe() {
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

	sub, err := client.Subscribe(ctx, appsync.SubscribeCommandInput{
		Channel: "/default/chat",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Close(ctx)

	for {
		var msg struct {
			Text string `json:"text"`
		}
		err := sub.DecodeNext(ctx, &msg)
		if errors.Is(err, appsync.ErrSubscriptionClosed) {
			return
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(msg.Text)
	}
}

func ExampleClient_Publish() {
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

	type message struct {
		Text string `json:"text"`
	}

	payload, err := json.Marshal(message{Text: "hello"})
	if err != nil {
		log.Fatal(err)
	}

	err = client.Publish(ctx, appsync.PublishCommandInput{
		Channel: "/default/chat",
		Payload: payload,
	})
	if err != nil {
		log.Fatal(err)
	}
}
