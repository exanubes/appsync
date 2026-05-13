package appsync

import (
	"context"

	"github.com/exanubes/appsync/authorizer"
)

// ConnectionOptions configures a WebSocket connection to an Appsync Events API.
type ConnectionOptions struct {
	Endpoint     string
	Subprotocols []string
	Authorizer   authorizer.Authorizer
	Backpressure Backpressure
}

// Backpressure controls the internal channel buffer sizes.
// Zero values are treated as "use the library default" (100 for each field) — setting a field
// to 0 does not produce an unbuffered channel; omit the field entirely to accept the default.
// Buffer sizes directly impact memory consumption: large values combined with many subscriptions
// or large payloads may result in substantial memory usage. Tune according to your workload.
// Limits are enforced per connection/subscription; no global memory cap is applied.
type Backpressure struct {
	ConnectionInbound  uint
	ConnectionOutbound uint
	SubscriptionEvents uint
}

type PublishCommandInput struct {
	Channel string
	Payload []byte
}

type SubscribeCommandInput struct {
	Channel string
}

type NextMessageOutput struct {
	Data []byte
}

// Client is the client-facing API for interacting with an AppSync WebSocket connection.
type Client interface {
	// Publish sends a message to a channel
	Publish(context.Context, PublishCommandInput) error
	// Subscribe to a channel and receive messages published to it
	Subscribe(context.Context, SubscribeCommandInput) (Subscription, error)
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
	// DecodeNext returns the next message from the channel and unmarshals it into value.
	// Blocks until a message is received or the context is cancelled.
	DecodeNext(context.Context, any) error
}
