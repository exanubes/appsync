package events

import (
	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
)

type ConnectionAckEvent struct {
	Type      string `json:"type"`
	TimeoutMs int    `json:"connectionTimeoutMs"`
}

func (event ConnectionAckEvent) ToProtocol() protocol.ConnectionAckMessage {
	return protocol.ConnectionAckMessage{
		TimeoutMs: event.TimeoutMs,
	}
}

type ErrorMetadata struct {
	Type    string `json:"errorType"`
	Message string `json:"message"`
}

type ErrorEvent struct {
	Type   string          `json:"type"`
	ID     string          `json:"id"`
	Errors []ErrorMetadata `json:"errors"`
}

func (event ErrorEvent) ToProtocol() protocol.ErrorMessage {
	errs := make([]protocol.ErrorMetadata, len(event.Errors))

	for index, err := range event.Errors {
		errs[index] = protocol.ErrorMetadata{
			Message: err.Message,
			Type:    err.Type,
		}
	}

	return protocol.ErrorMessage{
		ID:     event.ID,
		Errors: errs,
	}
}

type SuccessEvent struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func (event SuccessEvent) ToProtocol() protocol.SuccessMessage {
	return protocol.SuccessMessage{
		ID: event.ID,
	}
}

type DataEvent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data string `json:"event"`
}

func (event DataEvent) ToProtocol() protocol.DataMessage {
	return protocol.DataMessage{
		SubId:   event.ID,
		Payload: app.Payload(event.Data),
	}
}
