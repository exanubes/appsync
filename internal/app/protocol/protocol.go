package protocol

import "github.com/exanubes/appsync/internal/app"

type KeepAliveMessage struct{}

type ConnectionAckMessage struct {
	TimeoutMs int
}

type ErrorMessage struct {
	ID     string
	Errors []ErrorMetadata
}

type ErrorMetadata struct {
	Type    string
	Message string
}

type PublishMessage struct {
	Destination string
	Payload     app.Payload
}

type PublishResult struct {
	Err error
}

type SuccessMessage struct {
	ID string
}
