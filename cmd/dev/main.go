package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/exanubes/appsync"
)

var HTTP_ENDPOINT = os.Getenv("HTTP_ENDPOINT")
var WS_ENDPOINT = os.Getenv("WS_ENDPOINT")
var AWS_REGION = os.Getenv("AWS_REGION")
var CHANNEL = os.Getenv("CHANNEL")

func main() {
	println("HTTP_ENDPOINT", HTTP_ENDPOINT)
	println("WS_ENDPOINT", WS_ENDPOINT)
	println("AWS_REGION", AWS_REGION)
	println("CHANNEl", CHANNEL)
	ctx := context.Background()
	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		HttpEndpoint: HTTP_ENDPOINT,
		WsEndpoint:   WS_ENDPOINT,
		Region:       AWS_REGION,
		Subprotocols: []string{appsync.ProtocolEvents},
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
	err = client.Publish(ctx, appsync.PublishCommandInput{
		Payload: data,
		Channel: CHANNEL + "/test",
	})

	if err != nil {
		log.Fatal(err)
	}
}

type Event struct {
	Message string `json:"msg"`
}
