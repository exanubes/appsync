package codec

import "github.com/exanubes/appsync/internal/app"

type Codec struct{}

func New() *Codec {
	return &Codec{}
}

func (codec Codec) Encode(message app.Message, signature app.Signature) (app.Payload, error) {
	return nil, nil
}

func (codec Codec) Decode(payload app.Payload) (app.Message, error) {
	return nil, nil
}
