package data_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
	"github.com/exanubes/appsync/internal/app/subscription"
	"github.com/exanubes/appsync/internal/app/usecases/data"
)

type mock_registry struct {
	sub *subscription.Subscription
}

func (m *mock_registry) Get(_ string) *subscription.Subscription { return m.sub }

func active_sub() *subscription.Subscription {
	sub, _ := subscription.New("sub-id", "test-channel", 1)
	return sub
}

func inactive_sub() *subscription.Subscription {
	sub, _ := subscription.New("sub-id", "test-channel", 1)
	sub.Close()
	return sub
}

func full_sub() *subscription.Subscription {
	sub, _ := subscription.New("sub-id", "test-channel", 1)
	sub.Deliver(context.Background(), app.Payload("seed"))
	return sub
}

func TestReceiveData(t *testing.T) {
	const payload = "test-payload"

	tests := []struct {
		name           string
		registry       *mock_registry
		expect_err     error
		expect_payload bool
	}{
		{
			name:       "subscription not found",
			registry:   &mock_registry{sub: nil},
			expect_err: app.ErrSubscriptionNotFound,
		},
		{
			name:       "subscription is inactive",
			registry:   &mock_registry{sub: inactive_sub()},
			expect_err: app.ErrSubscriptionClosed,
		},
		{
			name:       "deliver error is propagated",
			registry:   &mock_registry{sub: full_sub()},
			expect_err: app.ErrSubscriptionInboxFull,
		},
		{
			name:           "successful delivery",
			registry:       &mock_registry{sub: active_sub()},
			expect_err:     nil,
			expect_payload: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := data.NewReceiveDataUsecase(tt.registry)

			err := usecase.Execute(context.Background(), protocol.DataMessage{
				SubId:   "sub-id",
				Payload: app.Payload(payload),
			})

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("got error %v, want %v", err, tt.expect_err)
			}

			if tt.expect_payload {
				received, err := tt.registry.sub.Next(context.Background())
				if err != nil {
					t.Fatalf("Next() returned unexpected error: %v", err)
				}
				if !bytes.Equal(received, []byte(payload)) {
					t.Errorf("got payload %q, want %q", received, payload)
				}
			}
		})
	}
}
