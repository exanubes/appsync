// Package appsync provides a Go client for the AWS AppSync Events API over WebSocket.
//
// It supports connecting to AppSync Event APIs, subscribing to channels,
// publishing events, and authorizing requests with API key, IAM, Lambda
// authorizer, Cognito User Pools, OIDC, or a custom authorizer.
//
// The two primary types are [Client], which manages the WebSocket connection,
// and [Subscription], which represents an active channel subscription. Start by
// calling [Connect] to obtain a Client, then use [Client.Subscribe] or
// [Client.Publish] to interact with channels.
//
// A single Client uses one authorizer for connection setup, subscribe, publish,
// and unsubscribe messages.
package appsync
