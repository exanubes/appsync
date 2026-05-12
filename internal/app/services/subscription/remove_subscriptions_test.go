package subscription_test

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"

	sub_service "github.com/exanubes/appsync/internal/app/services/subscription"
)

type mock_unsubscriber struct {
	mu           sync.Mutex
	err          error
	received_ids []string
}

func (m *mock_unsubscriber) Unsubscribe(_ context.Context, id string) error {
	m.mu.Lock()
	m.received_ids = append(m.received_ids, id)
	m.mu.Unlock()
	return m.err
}

func TestRemoveSubscriptions(t *testing.T) {
	sentinel_err := errors.New("unsubscribe failed")

	tests := []struct {
		name       string
		ids        []string
		mock_err   error
		expect_err error
	}{
		{
			name:       "no subscriptions",
			ids:        []string{},
			mock_err:   nil,
			expect_err: nil,
		},
		{
			name:       "success",
			ids:        []string{"a", "b", "c"},
			mock_err:   nil,
			expect_err: nil,
		},
		{
			name:       "error propagated",
			ids:        []string{"a"},
			mock_err:   sentinel_err,
			expect_err: sentinel_err,
		},
		{
			name:       "all errors joined",
			ids:        []string{"a", "b", "c"},
			mock_err:   sentinel_err,
			expect_err: sentinel_err,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mock_unsubscriber{err: tt.mock_err}
			service := sub_service.NewRemoveSubscriptionsService(mock)

			err := service.Remove(context.Background(), tt.ids...)

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("got error %v, want %v", err, tt.expect_err)
			}
		})
	}
}

func TestRemoveSubscriptions_ForwardsSubscriptionIds(t *testing.T) {
	ids := []string{"sub-1", "sub-2", "sub-3"}
	mock := &mock_unsubscriber{}
	service := sub_service.NewRemoveSubscriptionsService(mock)

	service.Remove(context.Background(), ids...)

	mock.mu.Lock()
	received := make([]string, len(mock.received_ids))
	copy(received, mock.received_ids)
	mock.mu.Unlock()

	if len(received) != len(ids) {
		t.Fatalf("received %d ids, want %d", len(received), len(ids))
	}

	sort.Strings(received)
	sort.Strings(ids)
	for i, id := range ids {
		if received[i] != id {
			t.Errorf("received_ids[%d] = %q, want %q", i, received[i], id)
		}
	}
}
