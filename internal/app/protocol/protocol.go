package protocol

import (
	"fmt"

	"github.com/exanubes/appsync/internal/app"
)

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

type DataMessage struct {
	SubId   string
	Payload app.Payload
}

type PublishResultMessage struct {
	ID       string
	Failures []FailedEvent
}

type FailedEvent struct {
	Success    bool
	Index      int
	RawMessage []byte
}

type BatchPublishError struct {
	Failures []FailedEvent
}

func (err BatchPublishError) Error() string {
	return fmt.Sprintf("batch publish failed: %d event(s) rejected", len(err.Failures))
}

func (err BatchPublishError) Is(target error) bool {
	_, ok := target.(BatchPublishError)
	return ok
}
