package internal

import (
	"context"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/exanubes/appsync/internal/app"
)

type CredentialProvider interface {
	Load(context.Context) (aws.Credentials, error)
}
type Signer interface {
	Sign(context.Context, *http.Request) (app.Signature, error)
}

type RequestFactory interface {
	Create(*url.URL, []byte) (*http.Request, error)
}
