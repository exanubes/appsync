package appsync

import (
	"context"

	"github.com/exanubes/appsync/logger"
	"github.com/exanubes/appsync/port"
)

type ConnectionOptions struct {
	Endpoint     string
	Subprotocols []string
	Authorizer   port.Authorizer
	Logger       logger.Logger
	Backpressure Backpressure
}

// Buffer sizes directly impact memory consumption.
// Large buffer configurations combined with high subscription counts or large payloads may result in substantial memory usage.
// Tune buffer sizes according to your workload characteristics and available system resources.
// Buffer limits are enforced per connection/subscription. The library does not impose a global memory limit or adaptive backpressure mechanism.
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
	// DecodeNext returns the next message from the channel and unmarshals it into value.
	// Blocks until a message is received or the context is cancelled.
	DecodeNext(context.Context, any) error
}
