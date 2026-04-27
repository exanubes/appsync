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
	"github.com/exanubes/appsync/internal/infrastructure/logger"
)

var HTTP_ENDPOINT = os.Getenv("HTTP_ENDPOINT")
var WS_ENDPOINT = os.Getenv("WS_ENDPOINT")
var AWS_REGION = os.Getenv("AWS_REGION")
var CHANNEL = os.Getenv("CHANNEL")
var APPSYNC_API_KEY = os.Getenv("APPSYNC_API_KEY")

func main() {
	println("HTTP_ENDPOINT", HTTP_ENDPOINT)
	println("WS_ENDPOINT", WS_ENDPOINT)
	println("AWS_REGION", AWS_REGION)
	println("CHANNEl", CHANNEL)
	println("APPSYNC_API_KEY", APPSYNC_API_KEY)
	ctx := context.Background()
	http_endpoint, _ := url.Parse(HTTP_ENDPOINT)
	// authorizer := appsync.NewIAMAuthorizer(AWS_REGION, http_endpoint)
	authorizer := appsync.NewApiKeyAuthorizer(APPSYNC_API_KEY, http_endpoint)
	logger := logger.New()
	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     WS_ENDPOINT,
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

	event := Event{
		Message: "Hello World!",
	}

	data, err := json.Marshal(event)

	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)

	defer cancel()

	output, err := client.Subscribe(ctx, appsync.SubscribeCommandInput{
		Channel: CHANNEL + "/test",
	})

	if err != nil {
		log.Fatal(err)
	}

	err = client.Publish(ctx, appsync.PublishCommandInput{
		Payload: data,
		Channel: CHANNEL + "/test",
	})

	if err != nil {
		log.Fatal(err)
	}
	ev := Event{}
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

type Event struct {
	Message string `json:"msg"`
}
