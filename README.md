[![Go Reference](https://pkg.go.dev/badge/github.com/exanubes/appsync.svg)](https://pkg.go.dev/github.com/exanubes/appsync)
[![Go Report Card](https://goreportcard.com/badge/github.com/exanubes/appsync)](https://goreportcard.com/report/github.com/exanubes/appsync)
[![License](https://img.shields.io/github/license/exanubes/appsync)](LICENSE)
[![Unit tests](https://github.com/exanubes/appsync/actions/workflows/tests.yaml/badge.svg)](https://github.com/exanubes/appsync/actions/workflows/tests.yaml)

# appsync — AWS AppSync Events WebSocket client for Go

`appsync` is a Go client library for the AWS AppSync Events WebSocket API.
It supports connecting to AppSync Event APIs, subscribing to channels, publishing events,
and authorizing requests with API key, IAM, Lambda authorizer, Cognito User Pools, OIDC, or a custom authorizer.

## Table of contents

- [Installation](#installation)
- [Core concepts](#core-concepts)
- [Endpoints](#endpoints)
- [Quick start](#quick-start)
- [Built-in authorizers](#built-in-authorizers)
  - [API key](#api-key)
  - [IAM](#iam)
  - [Token-based authorization](#token-based-authorization)
- [Publishing events](#publishing-events)
- [Subscribing to events](#subscribing-to-events)
- [Closing resources](#closing-resources)
- [Custom authorizers](#custom-authorizers)
- [Backpressure configuration](#backpressure-configuration)
- [Public errors](#public-errors)
- [Examples](#examples)
- [Limitations](#limitations)
- [Tips](#tips)
- [Status](#status)
- [License](#license)

## Installation

```bash
go get github.com/exanubes/appsync
```

```go
import (
	"github.com/exanubes/appsync"
	"github.com/exanubes/appsync/authorizer"
)
```

## Core concepts

The root package exposes two main abstractions:

```go
type Client interface {
    Publish(context.Context, PublishCommandInput) error
    Subscribe(context.Context, SubscribeCommandInput) (Subscription, error)
    Close(context.Context) error
}

type Subscription interface {
    Close(context.Context) error
    Next(context.Context) (*NextMessageOutput, error)
    DecodeNext(context.Context, any) error
}
```

Use `appsync.Connect` to establish one WebSocket connection. Use the returned `Client` to subscribe to channels and 
publish events. A `Subscription` receives events from one channel.

The AppSync Events WebSocket subprotocol value is exported as:

```go
appsync.ProtocolEvents // "aws-appsync-event-ws"
```

Pass it through `ConnectionOptions.Subprotocols` when connecting.

## Endpoints

The library uses two AppSync Events endpoints:

- the WebSocket realtime endpoint, used by the client to establish the connection;
- the HTTP event endpoint, used by authorizers to build the authorization headers expected by AppSync.

Example shape:

```go
httpEndpoint := "https://xxxxxxxxxxxxxxxxxxxx.appsync-api.us-east-1.amazonaws.com/event"
wsEndpoint := "wss://xxxxxxxxxxxxxxxxxxxx.appsync-realtime-api.us-east-1.amazonaws.com/event/realtime"
```



The authorizer uses the HTTP endpoint to create the authorization data expected by AppSync. The client uses the WebSocket endpoint to connect.

## Quick start

```go
func publish(ctx context.Context) error {
    httpEndpoint := "https://xxxxxxxxxxxxxxxxxxxx.appsync-api.us-east-1.amazonaws.com/event"
    wsEndpoint := "wss://xxxxxxxxxxxxxxxxxxxx.appsync-realtime-api.us-east-1.amazonaws.com/event/realtime"

	authz, err := authorizer.ApiKey(authorizer.ApiKeyAuthorizerConfig{
		ApiKey:   "your-api-key",
		Endpoint: httpEndpoint,
	})
	if err != nil {
		return err
	}

	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     wsEndpoint,
		Subprotocols: []string{appsync.ProtocolEvents},
		Authorizer:   authz,
	})
	if err != nil {
		return err
	}
	defer client.Close(context.Background())

	return client.Publish(ctx, appsync.PublishCommandInput{
		Channel: "default/notifications",
		Payload: []byte(`{"message":"hello"}`),
	})
}
```

## Built-in authorizers

The `authorizer` package includes built-in implementations for common AppSync Events authorization modes.

### API key

Use `authorizer.ApiKey` when your AppSync Events API is configured for API key authorization.

```go
authz, err := authorizer.ApiKey(authorizer.ApiKeyAuthorizerConfig{
    ApiKey:   "your-api-key",
    Endpoint: httpEndpoint,
})
if err != nil {
    return err
}
```

### IAM

Use `authorizer.IAM` when your AppSync Events API is configured for IAM authorization.

```go
authz, err := authorizer.IAM(authorizer.IAMAuthorizerConfig{
    Region:   "eu-central-1",
    Endpoint: httpEndpoint,
})
if err != nil {
    return err
}
```

IAM authorization uses the AWS SDK credential resolution. Environment credentials,
shared config/profile credentials, SSO, STS credentials, AssumeRole, ECS/EC2 role
credentials, and Lambda role credentials can be used when they are resolvable by
the AWS SDK configuration used by the authorizer.

### Token-based authorization

Use `authorizer.Token` for AppSync authorization modes where AppSync expects an `Authorization` token, including Lambda 
authorizers, Cognito User Pool tokens, and OpenID Connect tokens.

```go
authz, err := authorizer.Token(authorizer.TokenAuthorizerConfig{
    AuthToken: token,
    Endpoint:  httpEndpoint,
})
if err != nil {
    return err
}
```

Examples:

```go
// Lambda authorizer token
authz, err := authorizer.Token(authorizer.TokenAuthorizerConfig{
    AuthToken: "custom-token",
    Endpoint:  httpEndpoint,
})

// Cognito User Pool ID token
authz, err := authorizer.Token(authorizer.TokenAuthorizerConfig{
    AuthToken: cognitoIDToken,
    Endpoint:  httpEndpoint,
})

// OIDC token
authz, err := authorizer.Token(authorizer.TokenAuthorizerConfig{
    AuthToken: oidcToken,
    Endpoint:  httpEndpoint,
})
```

## Publishing events

`Publish` sends a payload to a channel.

```go
payload := []byte(`{"message":"hello"}`)

err := client.Publish(ctx, appsync.PublishCommandInput{
    Channel: "default/notifications",
    Payload: payload,
})
if err != nil {
    return err
}
```

`Payload` is a raw byte slice. The library does not require a Go struct, but AppSync event payloads are commonly JSON. 
If you want structured data, marshal it before publishing.

## Subscribing to events

Use `Subscribe` to create a channel subscription.

```go
sub, err := client.Subscribe(ctx, appsync.SubscribeCommandInput{
    Channel: "default/notifications",
})
if err != nil {
    return err
}
defer sub.Close(context.Background())
```

Read event messages with `Next`:

```go
message, err := sub.Next(ctx)
if err != nil {
    return err
}

fmt.Printf("raw payload: %s\n", message.Data)
```

Decode JSON payloads with `DecodeNext`:

```go
type Notification struct {
    Message string `json:"message"`
}

var notification Notification
if err := sub.DecodeNext(ctx, &notification); err != nil {
    return err
}
```

`Next` and `DecodeNext` block until one of these happens:

- a message arrives,
- the context is cancelled or reaches its deadline,
- the subscription is closed.

Use context deadlines on read operations if the caller cannot block indefinitely.

## Closing resources

Close subscriptions when you no longer need channel events:

```go
if err := sub.Close(ctx); err != nil {
    return err
}
```

Close the client when the WebSocket connection is no longer needed:

```go
if err := client.Close(ctx); err != nil {
    return err
}
```

Closing the client closes the WebSocket connection and all open subscriptions created on it.

A closed subscription returns `appsync.ErrSubscriptionClosed` from later reads. Calling `Close` on an already closed 
subscription can also return `appsync.ErrSubscriptionClosed`.

## Custom authorizers

Implement `authorizer.Authorizer` when the built-in authorizers do not fit your authorization model.

```go
type Authorizer interface {
    Authorize(context.Context, AuthorizeCommandInput) (*AuthorizeCommandOutput, error)
}

type AuthorizeCommandInput struct {
    Channel string
    Payload []byte
}

type AuthorizeCommandOutput struct {
    Signature map[string]string
}
```

The returned `Signature` map should contain the authorization fields AppSync expects. The library uses that map for the 
WebSocket connection handshake and for outgoing subscribe, publish, and unsubscribe operations.

A single `Client` uses the same authorizer for all of those operations.

Important: `Authorize` must handle empty input. The library calls it in multiple situations:

| Operation         | `Channel`            | `Payload`       |
|-------------------|----------------------|-----------------|
| Connect handshake | empty                | nil             |
| Subscribe         | subscription channel | nil             |
| Publish           | destination channel  | publish payload |
| Unsubscribe       | empty                | nil             |

A minimal static custom authorizer can look like this:

```go
package main

import (
    "context"
    "net/url"

    "github.com/exanubes/appsync/authorizer"
)

type StaticAuthorizer struct {
    token string
    host  string
}

func NewStaticAuthorizer(endpoint string, token string) (*StaticAuthorizer, error) {
    parsed, err := url.Parse(endpoint)
    if err != nil {
        return nil, err
    }

    return &StaticAuthorizer{
        token: token,
        host:  parsed.Host,
    }, nil
}

func (authz *StaticAuthorizer) Authorize(
    ctx context.Context,
    input authorizer.AuthorizeCommandInput,
) (*authorizer.AuthorizeCommandOutput, error) {
    return &authorizer.AuthorizeCommandOutput{
        Signature: map[string]string{
            "Authorization": authz.token,
            "host":          authz.host,
        },
    }, nil
}
```

Then pass it to `appsync.Connect`:

```go
authz, err := NewStaticAuthorizer(httpEndpoint, token)
if err != nil {
    return err
}

client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
    Endpoint:     wsEndpoint,
    Subprotocols: []string{appsync.ProtocolEvents},
    Authorizer:   authz,
})
```

### Custom authorizer with per-message signing

Some authorization schemes need the channel and payload to compute a signature. `AuthorizeCommandInput` exposes both values.

```go
func (authz *SigningAuthorizer) Authorize(
    ctx context.Context,
    input authorizer.AuthorizeCommandInput,
) (*authorizer.AuthorizeCommandOutput, error) {
    signature, err := authz.sign(ctx, input.Channel, input.Payload)
    if err != nil {
        return nil, err
    }

    return &authorizer.AuthorizeCommandOutput{
        Signature: map[string]string{
            "Authorization": signature,
            "host":          authz.host,
        },
    }, nil
}
```

For connection and unsubscribe calls, `input.Channel` is empty and `input.Payload` is nil. The signing function must treat that as a valid case.

## Backpressure configuration

`ConnectionOptions.Backpressure` controls internal buffer sizes.

```go
client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
    Endpoint:     wsEndpoint,
    Subprotocols: []string{appsync.ProtocolEvents},
    Authorizer:   authz,
    Backpressure: appsync.Backpressure{
        ConnectionInbound:  100,
        ConnectionOutbound: 100,
        SubscriptionEvents: 100,
    },
})
```

Fields:

| Field                | Meaning                                                                    |
|----------------------|----------------------------------------------------------------------------|
| `ConnectionInbound`  | Buffer for messages received from the WebSocket connection before routing. |
| `ConnectionOutbound` | Buffer for messages waiting to be written to the WebSocket connection.     |
| `SubscriptionEvents` | Buffer for events waiting to be consumed by a subscription.                |

Zero values use the library default of `100`. Setting a field to `0` does not create an unbuffered channel.

Large buffers can increase memory usage, especially with many subscriptions or large payloads. There is no global 
memory cap exposed by the public API.

If a subscription event buffer stays full, event delivery can fail with `appsync.ErrSubscriptionInboxFull`.

## Public errors

The root package exposes sentinel errors that callers can check with `errors.Is`:

- `appsync.ErrEmptyUrl`
- `appsync.ErrHandshakeTimeout`
- `appsync.ErrDuplicateMessage`
- `appsync.ErrSubscriptionInboxFull`
- `appsync.ErrSubscriptionClosed`
- `appsync.ErrSubscriptionNotFound`
- `appsync.ErrHeartbeatTimeout`

Typical handling:

```go
message, err := sub.Next(ctx)
if err != nil {
    switch {
    case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
        return err
    case errors.Is(err, appsync.ErrSubscriptionClosed):
        return nil
    default:
        return err
    }
}

fmt.Printf("received: %s\n", message.Data)
```

Relevant behavior:

- `ErrHandshakeTimeout` means the WebSocket connection was opened, but AppSync did not acknowledge the connection initialization in time.
- `ErrHeartbeatTimeout` means keep-alive messages stopped arriving within the expected connection timeout window.
- `ErrSubscriptionClosed` means the subscription is no longer active.
- `ErrSubscriptionInboxFull` means the subscriber did not consume events fast enough for its configured buffer.
- Context cancellation and deadlines are propagated from public methods where applicable.

## Examples

Runnable examples are available in:

- [`examples/api-key`](examples/api-key)
- [`examples/iam`](examples/iam)
- [`examples/token`](examples/token)
- [`examples/custom-authorizer`](examples/custom-authorizer)

## Limitations

A `Client` uses one authorizer for the entire connection lifecycle.

The same authorizer is used to establish the WebSocket connection and to authorize
`subscribe`, `publish`, and `unsubscribe` messages. Using different authorizers for
connection setup and individual operation messages is not currently supported.

## Tips

### Use one client per WebSocket connection

A `Client` represents one active AppSync WebSocket connection. Multiple subscriptions can be created from the same client.

### Always close what you open

Close individual subscriptions when a channel is no longer needed. Close the client when shutting down the process or 
component that owns the connection.

### Put deadlines on blocking calls

`Connect`, `Publish`, `Subscribe`, `Subscription.Close`, `Subscription.Next`, `Subscription.DecodeNext`, and 
`Client.Close` all accept `context.Context`. Use deadlines when the caller has a bounded lifecycle.

### Treat payloads as application-owned bytes

The library accepts and returns payloads as `[]byte`. It does not impose an application schema. Use `json.Marshal` and `DecodeNext` 
when your event contract is JSON.

### Keep custom authorizers side-effect safe

A custom authorizer may be called for every connect, subscribe, publish, and unsubscribe operation. Avoid expensive 
work where possible, cache stable data safely, and refresh credentials/tokens deliberately when your auth model 
requires it.

## Status

The API is the desired shape and it "works for me", however, since the library does not yet support the full Appsync 
Events API featureset, I've decided to have it as a v0 in case somebody actually uses this and I need to break the API
in the future for some reason.

Missing features:

- authorizer per request
- HTTP Publish
- Batch Publish 
- something else I missed probably

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

