// Package appsync provides a client for the AWS AppSync Events API over WebSocket.
//
// The two primary types are [Client], which manages the WebSocket connection, and
// [Subscription], which represents an active channel subscription. Start by calling
// [Connect] to obtain a Client, then use Subscribe or Publish to interact with channels.
package appsync
