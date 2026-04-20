package appsync

import (
	"context"
)

type ConnectionOptions struct {
	HttpEndpoint string
	WsEndpoint   string
	Subprotocols []string
	Region       string
	Endpoint     string
}

type PublishCommandInput struct {
	Channel string
	Payload []byte
}

type SubscribeCommandInput struct{}
type SubscribeCommandOutput struct {
	Sub Subscription
}

type NextMessageOutput struct {
	Data []byte
}

// Client is the client-facing API for interacting with an AppSync WebSocket connection.
type Client interface {
	// Send a publish message via websocket connection to a particular channel
	Publish(context.Context, PublishCommandInput) error
	// Subscribe to a channel and receive messages published to it
	Subscribe(context.Context, SubscribeCommandInput) (*SubscribeCommandOutput, error)
	// Close the websocket connection and all open subscriptions created on it
	Close(context.Context) error
}

// Subscription represents an active subscription to a channel.
type Subscription interface {
	// Close stops receiving messages on the channel.
	Close(context.Context) error
	// Next returns the next message from the channel.
	// Blocks until a message is received or the context is cancelled.
	Next(context.Context) (*NextMessageOutput, error)
	// Decode returns the next message from the channel and unmarshals it into value.
	// Blocks until a message is received or the context is cancelled.
	Decode(context.Context, any) error
}
