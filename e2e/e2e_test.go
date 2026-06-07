//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	"errors"
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
			namespace: require_env(t, "APPSYNC_E2E_NS_API_KEY"),
			authz: func(endpoint string) (authorizer.Authorizer, error) {
				return authorizer.ApiKey(authorizer.ApiKeyAuthorizerConfig{
					ApiKey:   require_env(t, "APPSYNC_E2E_API_KEY"),
					Endpoint: endpoint,
				})
			},
		},
		{
			name:      "iam",
			namespace: require_env(t, "APPSYNC_E2E_NS_IAM"),
			authz: func(endpoint string) (authorizer.Authorizer, error) {
				return authorizer.IAM(authorizer.IAMAuthorizerConfig{
					Region:   aws_region,
					Endpoint: endpoint,
				})
			},
		},
		{
			name:      "lambda",
			namespace: require_env(t, "APPSYNC_E2E_NS_LAMBDA"),
			authz: func(endpoint string) (authorizer.Authorizer, error) {
				return authorizer.Token(authorizer.TokenAuthorizerConfig{
					AuthToken: require_env(t, "APPSYNC_E2E_LAMBDA_TOKEN"),
					Endpoint:  endpoint,
				})
			},
		},
		{
			name:      "cognito",
			namespace: require_env(t, "APPSYNC_E2E_NS_COGNITO"),
			authz: func(endpoint string) (authorizer.Authorizer, error) {
				return authorizer.Token(authorizer.TokenAuthorizerConfig{
					AuthToken: require_env(t, "APPSYNC_E2E_COGNITO_ID_TOKEN"),
					Endpoint:  endpoint,
				})
			},
		},
		{
			name:      "oidc",
			namespace: require_env(t, "APPSYNC_E2E_NS_OIDC"),
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
				Authorizers:  appsync.Authorizers{Default: authz},
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

func TestSuccessfulUnsubscribe(t *testing.T) {
	http_endpoint := require_env(t, "APPSYNC_E2E_HTTP_ENDPOINT")
	ws_endpoint := require_env(t, "APPSYNC_E2E_WS_ENDPOINT")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	authz, err := authorizer.ApiKey(authorizer.ApiKeyAuthorizerConfig{
		ApiKey:   require_env(t, "APPSYNC_E2E_API_KEY"),
		Endpoint: http_endpoint,
	})
	if err != nil {
		t.Fatalf("create authorizer: %v", err)
	}

	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     ws_endpoint,
		Subprotocols: []string{appsync.ProtocolEvents},
		Authorizers:  appsync.Authorizers{Default: authz},
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

	namespace := require_env(t, "APPSYNC_E2E_NS_API_KEY")
	channel := fmt.Sprintf("%s/test-%d", namespace, time.Now().UnixNano())

	sub, err := client.Subscribe(ctx, appsync.SubscribeCommandInput{
		Channel: channel,
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	close_ctx, close_cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer close_cancel()

	if err := sub.Close(close_ctx); err != nil {
		t.Fatalf("unsubscribe: %v", err)
	}

	_, err = sub.Next(ctx)
	if !errors.Is(err, appsync.ErrSubscriptionClosed) {
		t.Fatalf("expected ErrSubscriptionClosed after unsubscribe, got: %v", err)
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
		Authorizers:  appsync.Authorizers{Default: authz},
	})

	if err == nil {
		t.Fatal("connect succeeded with invalid API key")
	}
}

func TestPerRequestAuthorizerOverride(t *testing.T) {
	http_endpoint := require_env(t, "APPSYNC_E2E_HTTP_ENDPOINT")
	ws_endpoint := require_env(t, "APPSYNC_E2E_WS_ENDPOINT")
	aws_region := require_env(t, "AWS_REGION")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	apiKey, err := authorizer.ApiKey(authorizer.ApiKeyAuthorizerConfig{
		ApiKey:   require_env(t, "APPSYNC_E2E_API_KEY"),
		Endpoint: http_endpoint,
	})
	if err != nil {
		t.Fatalf("create api_key authorizer: %v", err)
	}

	iam, err := authorizer.IAM(authorizer.IAMAuthorizerConfig{
		Region:   aws_region,
		Endpoint: http_endpoint,
	})
	if err != nil {
		t.Fatalf("create iam authorizer: %v", err)
	}

	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     ws_endpoint,
		Subprotocols: []string{appsync.ProtocolEvents},
		Authorizers:  appsync.Authorizers{Default: apiKey},
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

	namespace := require_env(t, "APPSYNC_E2E_NS_IAM")
	channel := fmt.Sprintf("%s/test-%d", namespace, time.Now().UnixNano())
	payload := []byte(`{"test":"per_request_override"}`)

	sub, err := client.Subscribe(ctx, appsync.SubscribeCommandInput{
		Channel:    channel,
		Authorizer: iam,
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
		Channel:    channel,
		Payload:    payload,
		Authorizer: iam,
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
}

func TestConnectionSpecificAuthorizers(t *testing.T) {
	http_endpoint := require_env(t, "APPSYNC_E2E_HTTP_ENDPOINT")
	ws_endpoint := require_env(t, "APPSYNC_E2E_WS_ENDPOINT")
	aws_region := require_env(t, "AWS_REGION")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	iam, err := authorizer.IAM(authorizer.IAMAuthorizerConfig{
		Region:   aws_region,
		Endpoint: http_endpoint,
	})
	if err != nil {
		t.Fatalf("create iam authorizer: %v", err)
	}

	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     ws_endpoint,
		Subprotocols: []string{appsync.ProtocolEvents},
		Authorizers: appsync.Authorizers{
			Connect:   iam,
			Subscribe: iam,
			Publish:   iam,
		},
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

	namespace := require_env(t, "APPSYNC_E2E_NS_IAM")
	channel := fmt.Sprintf("%s/test-%d", namespace, time.Now().UnixNano())
	payload := []byte(`{"test":"connection_specific_authorizers"}`)

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
}

func TestMixedConnectionLevelAuthorizers(t *testing.T) {
	http_endpoint := require_env(t, "APPSYNC_E2E_HTTP_ENDPOINT")
	ws_endpoint := require_env(t, "APPSYNC_E2E_WS_ENDPOINT")
	aws_region := require_env(t, "AWS_REGION")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	apiKey, err := authorizer.ApiKey(authorizer.ApiKeyAuthorizerConfig{
		ApiKey:   require_env(t, "APPSYNC_E2E_API_KEY"),
		Endpoint: http_endpoint,
	})
	if err != nil {
		t.Fatalf("create api_key authorizer: %v", err)
	}

	iam, err := authorizer.IAM(authorizer.IAMAuthorizerConfig{
		Region:   aws_region,
		Endpoint: http_endpoint,
	})
	if err != nil {
		t.Fatalf("create iam authorizer: %v", err)
	}

	client, err := appsync.Connect(ctx, appsync.ConnectionOptions{
		Endpoint:     ws_endpoint,
		Subprotocols: []string{appsync.ProtocolEvents},
		Authorizers: appsync.Authorizers{
			Connect:   iam,
			Subscribe: apiKey,
			Publish:   apiKey,
		},
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

	namespace := require_env(t, "APPSYNC_E2E_NS_API_KEY")
	channel := fmt.Sprintf("%s/test-%d", namespace, time.Now().UnixNano())
	payload := []byte(`{"test":"mixed_connection_level_authorizers"}`)

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
}

func require_env(t *testing.T, key string) string {
	t.Helper()

	value := os.Getenv(key)
	if value == "" {
		t.Fatalf("%s is required", key)
	}

	return value
}
