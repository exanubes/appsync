package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/exanubes/appsync"
	"github.com/exanubes/appsync/authorizer"
	"github.com/exanubes/appsync/internal/infrastructure/logger"
)

var http_endpoint = os.Getenv("HTTP_ENDPOINT")
var ws_endpoint = os.Getenv("WS_ENDPOINT")
var aws_region = os.Getenv("AWS_REGION")
var channel = os.Getenv("CHANNEL")
var appsync_api_key = os.Getenv("APPSYNC_API_KEY")
var cognito_auth_token = os.Getenv("ID_TOKEN")
var oidc_auth_token = os.Getenv("OIDC_TOKEN")

func main() {
	println("HTTP_ENDPOINT", http_endpoint)
	println("WS_ENDPOINT", ws_endpoint)
	println("AWS_REGION", aws_region)
	println("CHANNEl", channel)
	println("APPSYNC_API_KEY", appsync_api_key)
	ctx := context.Background()
	http_endpoint, _ := url.Parse(http_endpoint)
	authorizer := authorizer.IAM(aws_region, http_endpoint)
	// authorizer := authorizer.ApiKey(APPSYNC_API_KEY, http_endpoint)
	// authorizer := authorizer.Token("custom-token", http_endpoint)
	// authorizer := authorizer.Token(COGNITO_AUTH_TOKEN, http_endpoint)
	// authorizer := authorizer.Token(OIDC_AUTH_TOKEN, http_endpoint)
	logger := logger.New()
	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     ws_endpoint,
		Subprotocols: []string{appsync.ProtocolEvents},
		Authorizer:   authorizer,
		Logger:       logger,
	})

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := client.Close(ctx)
		if err != nil {
			println("Closed with error: ", err.Error())

		} else {
			println("Closed without error")
		}
	}()

	event := event_msg{
		Message: "Hello World!",
	}

	data, err := json.Marshal(event)

	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)

	defer cancel()

	output, err := client.Subscribe(ctx, appsync.SubscribeCommandInput{
		Channel: channel + "/test",
	})

	if err != nil {
		log.Fatal(err)
	}

	err = client.Publish(ctx, appsync.PublishCommandInput{
		Payload: data,
		Channel: channel + "/test",
	})

	if err != nil {
		log.Fatal(err)
	}
	ev := event_msg{}
	err = output.Sub.DecodeNext(ctx, &ev)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("EVENT: ", ev)

	err = output.Sub.Close(ctx)

	if err != nil {
		println("Unsub err: ", err.Error())
	}

	err = output.Sub.Close(ctx)

	if err != nil {
		println("Second unsub err: ", err.Error())
	}

	_, err = output.Sub.Next(ctx)

	if err != nil {
		println("Next after close err: ", err.Error())
	}
}

type event_msg struct {
	Message string `json:"msg"`
}
