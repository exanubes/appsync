package connection

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type GenerateSubprotocolService struct {
	authorizer app.RequestAuthorizer
	serializer Serializer
}

func NewGenerateSubprotocolService(authorizer app.RequestAuthorizer, serializer Serializer) *GenerateSubprotocolService {
	return &GenerateSubprotocolService{authorizer, serializer}
}

func (service *GenerateSubprotocolService) Generate(ctx context.Context) (string, error) {
	signature, err := service.authorizer.Authorize(ctx, []byte(`{}`))
	if err != nil {
		return "", err
	}

	encoded, err := service.serializer.Serialize(signature)

	if err != nil {
		return "", err
	}

	return "header-" + encoded, nil

}
