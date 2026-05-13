package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/exanubes/appsync"
	"github.com/exanubes/appsync/authorizer"
)

func requiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}

// refreshingAuthorizer fetches a fresh token before each authorization call.
// This is useful when tokens are short-lived (e.g. Cognito, STS, custom auth servers).
type refreshingAuthorizer struct {
	host    string
	authURL string
	client  *http.Client
}

func (a *refreshingAuthorizer) Authorize(ctx context.Context, _ authorizer.AuthorizeCommandInput) (*authorizer.AuthorizeCommandOutput, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.authURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Token string `json:"token"`
	}
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &authorizer.AuthorizeCommandOutput{
		Signature: map[string]string{
			"Authorization": result.Token,
			"host":          a.host,
		},
	}, nil
}

func main() {
	var (
		httpEndpoint = requiredEnv("APPSYNC_HTTP_ENDPOINT")
		wsEndpoint   = requiredEnv("APPSYNC_WS_ENDPOINT")
		authURL      = requiredEnv("AUTH_URL")
		channel      = requiredEnv("APPSYNC_CHANNEL")
	)

	parsed, err := url.Parse(httpEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	authz := &refreshingAuthorizer{
		host:    parsed.Host,
		authURL: authURL,
		client:  &http.Client{},
	}

	ctx := context.Background()

	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     wsEndpoint,
		Subprotocols: []string{appsync.ProtocolEvents},
		Authorizer:   authz,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close(ctx)

	sub, err := client.Subscribe(ctx, appsync.SubscribeCommandInput{
		Channel: channel,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Close(ctx)

	type event struct {
		Message string `json:"message"`
	}

	payload, err := json.Marshal(event{Message: "hello from custom-authorizer example"})
	if err != nil {
		log.Fatal(err)
	}

	if err = client.Publish(ctx, appsync.PublishCommandInput{
		Channel: channel,
		Payload: payload,
	}); err != nil {
		log.Fatal(err)
	}

	var received event
	if err = sub.DecodeNext(ctx, &received); err != nil {
		if errors.Is(err, appsync.ErrSubscriptionClosed) {
			return
		}
		log.Fatal(err)
	}

	fmt.Println("received:", received.Message)
}
