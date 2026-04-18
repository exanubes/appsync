package main

import (
	"context"
	"log"
	"os"

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
	_, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		HttpEndpoint: HTTP_ENDPOINT,
		WsEndpoint:   WS_ENDPOINT,
		Region:       AWS_REGION,
		Subprotocols: []string{appsync.ProtocolEvents},
	})

	if err != nil {
		log.Fatal(err)
	}
}
