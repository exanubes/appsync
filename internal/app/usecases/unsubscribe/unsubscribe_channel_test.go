package unsubscribe_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app/usecases/unsubscribe"
)

type mock_unsubscriber struct {
	err         error
	called      bool
	received_id string
}

func (m *mock_unsubscriber) Unsubscribe(_ context.Context, id string) error {
	m.called = true
	m.received_id = id
	return m.err
}

func TestUnsubscribeChannel(t *testing.T) {
	const subscriptionId = "sub-id"
	sentinel_err := errors.New("unsubscribe failed")

	tests := []struct {
		name       string
		mock       *mock_unsubscriber
		expect_err error
	}{
		{
			name:       "success",
			mock:       &mock_unsubscriber{},
			expect_err: nil,
		},
		{
			name:       "error propagated",
			mock:       &mock_unsubscriber{err: sentinel_err},
			expect_err: sentinel_err,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := unsubscribe.NewUnsubscribeChannelUsecase(tt.mock)

			err := usecase.Execute(context.Background(), unsubscribe.UnsubscribeChannelCommandInput{
				SubscriptionId: subscriptionId,
			})

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("got error %v, want %v", err, tt.expect_err)
			}
			if !tt.mock.called {
				t.Error("Unsubscribe was not called")
			}
		})
	}
}

func TestUnsubscribeChannel_ForwardsSubscriptionId(t *testing.T) {
	const subscriptionId = "my-sub-id"
	mock := &mock_unsubscriber{}
	usecase := unsubscribe.NewUnsubscribeChannelUsecase(mock)

	usecase.Execute(context.Background(), unsubscribe.UnsubscribeChannelCommandInput{
		SubscriptionId: subscriptionId,
	})

	if mock.received_id != subscriptionId {
		t.Errorf("received_id = %q, want %q", mock.received_id, subscriptionId)
	}
}
