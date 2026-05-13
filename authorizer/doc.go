// Package authorizer provides built-in authorizers for AWS AppSync Events.
//
// It includes API key, IAM, and token-based authorizers. Token-based
// authorization can be used with Lambda authorizers, Cognito User Pool tokens,
// and OpenID Connect tokens.
//
// Custom authorization schemes can implement [Authorizer].
package authorizer
