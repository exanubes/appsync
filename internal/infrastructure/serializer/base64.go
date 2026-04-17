package serializer

import (
	"encoding/base64"
	"encoding/json"

	"github.com/exanubes/appsync/internal/app"
)

type Base64Serializer struct{}

func New() *Base64Serializer {
	return &Base64Serializer{}
}

func (*Base64Serializer) Serialize(signature app.Signature) (string, error) {
	data, err := json.Marshal(signature)

	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}
