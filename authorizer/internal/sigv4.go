package internal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/exanubes/appsync/internal/app"
)

type Sigv4Signer struct {
	Provider CredentialProvider
	Region   string
	Clock    app.Clock
}

func (signer *Sigv4Signer) Sign(ctx context.Context, req *http.Request) (app.Signature, error) {
	payload, err := read_body(req)

	if err != nil {
		return app.Signature{}, err
	}

	sigv4 := v4.NewSigner()
	credentials, err := signer.Provider.Load(ctx)
	if err != nil {
		return app.Signature{}, err
	}

	sum := sha256.Sum256(payload)
	payload_hash := hex.EncodeToString(sum[:])

	err = sigv4.SignHTTP(
		ctx,
		credentials,
		req,
		payload_hash,
		"appsync",
		signer.Region,
		signer.Clock.Now(),
	)

	if err != nil {
		return app.Signature{}, err
	}

	return parse_headers(req), err
}

func read_body(req *http.Request) ([]byte, error) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return []byte(`{}`), nil
	}

	return body, nil
}

func parse_headers(req *http.Request) app.Signature {
	keys := []string{"Authorization", "accept", "content-encoding", "content-type", "host", "x-amz-date", "x-amz-security-token"}
	headers := map[string]string{}

	for _, key := range keys {
		value := req.Header.Get(key)
		if value != "" {
			headers[key] = value
		}
	}

	return headers
}
