package codec

import (
	"encoding/json"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
	"github.com/exanubes/appsync/internal/infrastructure/events"
)

type Envelope struct {
	Type string `json:"type"`
}
type Codec struct{}

func New() *Codec {
	return &Codec{}
}

func (codec Codec) Encode(msg app.Message) (app.Payload, error) {
	return json.Marshal(msg)
}

func (codec Codec) Decode(payload app.Payload) (app.Message, error) {
	var msg Envelope
	err := json.Unmarshal(payload, &msg)

	switch msg.Type {
	case protocol.TypeConnectionAck:
		event := events.ConnectionAckEvent{}
		err := json.Unmarshal(payload, &event)
		return event.ToProtocol(), err
	case protocol.TypeKeepAlive:
		return protocol.KeepAliveMessage{}, nil

	case protocol.TypePublishSuccess:
		event := events.SuccessEvent{}
		err := json.Unmarshal(payload, &event)
		return event.ToProtocol(), err
	case protocol.TypeError, protocol.TypePublishError:
		event := events.ErrorEvent{}
		err := json.Unmarshal(payload, &event)
		return event.ToProtocol(), err
	}

	return msg, err
}
