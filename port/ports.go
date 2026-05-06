package port

import "context"

// Authorizer is used for generating subprotocol and authorizing outgoing messages
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
