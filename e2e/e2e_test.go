//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/exanubes/appsync"
	"github.com/exanubes/appsync/authorizer"
)

type auth_case struct {
	name      string
	namespace string
	authz     func(string) (authorizer.Authorizer, error)
}

func TestAppSyncAuthorizers(t *testing.T) {
	http_endpoint := require_env(t, "APPSYNC_E2E_HTTP_ENDPOINT")
	ws_endpoint := require_env(t, "APPSYNC_E2E_WS_ENDPOINT")
	aws_region := require_env(t, "AWS_REGION")

	cases := []auth_case{
		{
			name:      "api_key",
			namespace: "api-key-e2e",
			authz: func(endpoint string) (authorizer.Authorizer, error) {
				return authorizer.ApiKey(authorizer.ApiKeyAuthorizerConfig{
					ApiKey:   require_env(t, "APPSYNC_E2E_API_KEY"),
					Endpoint: endpoint,
				})
			},
		},
		{
			name:      "iam",
			namespace: "iam-e2e",
			authz: func(endpoint string) (authorizer.Authorizer, error) {
				return authorizer.IAM(authorizer.IAMAuthorizerConfig{
					Region:   aws_region,
					Endpoint: endpoint,
				})
			},
		},
		{
			name:      "lambda",
			namespace: "lambda-e2e",
			authz: func(endpoint string) (authorizer.Authorizer, error) {
				return authorizer.Token(authorizer.TokenAuthorizerConfig{
					AuthToken: require_env(t, "APPSYNC_E2E_LAMBDA_TOKEN"),
					Endpoint:  endpoint,
				})
			},
		},
		{
			name:      "cognito",
			namespace: "cognito-e2e",
			authz: func(endpoint string) (authorizer.Authorizer, error) {
				return authorizer.Token(authorizer.TokenAuthorizerConfig{
					AuthToken: require_env(t, "APPSYNC_E2E_COGNITO_ID_TOKEN"),
					Endpoint:  endpoint,
				})
			},
		},
		{
			name:      "oidc",
			namespace: "oidc-e2e",
			authz: func(endpoint string) (authorizer.Authorizer, error) {
				return authorizer.Token(authorizer.TokenAuthorizerConfig{
					AuthToken: require_env(t, "APPSYNC_E2E_OIDC_TOKEN"),
					Endpoint:  endpoint,
				})
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			authz, err := tc.authz(http_endpoint)
			if err != nil {
				t.Fatalf("create authorizer: %v", err)
			}

			client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
				Endpoint:     ws_endpoint,
				Subprotocols: []string{appsync.ProtocolEvents},
				Authorizer:   authz,
			})
			if err != nil {
				t.Fatalf("connect: %v", err)
			}

			defer func() {
				close_ctx, close_cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer close_cancel()

				if err := client.Close(close_ctx); err != nil {
					t.Logf("client close: %v", err)
				}
			}()

			channel := fmt.Sprintf("%s/test-%d", tc.namespace, time.Now().UnixNano())
			payload := []byte(fmt.Sprintf(`{"authorizer":%q,"channel":%q}`, tc.name, channel))

			sub, err := client.Subscribe(ctx, appsync.SubscribeCommandInput{
				Channel: channel,
			})
			if err != nil {
				t.Fatalf("subscribe: %v", err)
			}

			defer func() {
				close_ctx, close_cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer close_cancel()

				if err := sub.Close(close_ctx); err != nil {
					t.Logf("subscription close: %v", err)
				}
			}()

			if err := client.Publish(ctx, appsync.PublishCommandInput{
				Channel: channel,
				Payload: payload,
			}); err != nil {
				t.Fatalf("publish: %v", err)
			}

			message, err := sub.Next(ctx)
			if err != nil {
				t.Fatalf("next message: %v", err)
			}

			if !bytes.Equal(message.Data, payload) {
				t.Fatalf("message payload mismatch\nwant: %s\n got: %s", payload, message.Data)
			}
		})
	}
}

func TestInvalidApiKeyCannotConnect(t *testing.T) {
	http_endpoint := require_env(t, "APPSYNC_E2E_HTTP_ENDPOINT")
	ws_endpoint := require_env(t, "APPSYNC_E2E_WS_ENDPOINT")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	authz, err := authorizer.ApiKey(authorizer.ApiKeyAuthorizerConfig{
		ApiKey:   "invalid-api-key",
		Endpoint: http_endpoint,
	})
	if err != nil {
		t.Fatalf("create authorizer: %v", err)
	}

	_, err = appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     ws_endpoint,
		Subprotocols: []string{appsync.ProtocolEvents},
		Authorizer:   authz,
	})

	if err == nil {
		t.Fatal("connect succeeded with invalid API key")
	}
}

func require_env(t *testing.T, key string) string {
	t.Helper()

	value := os.Getenv(key)
	if value == "" {
		t.Fatalf("%s is required", key)
	}

	return value
}
