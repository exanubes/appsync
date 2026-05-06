package subscription_test

import (
	"testing"

	sub_service "github.com/exanubes/appsync/internal/app/services/subscription"
	"github.com/exanubes/appsync/internal/app/subscription"
)

type mock_registry struct {
	registered_sub *subscription.Subscription
	called         bool
}

func (mr *mock_registry) Register(sub *subscription.Subscription) {
	mr.called = true
	mr.registered_sub = sub
}

func TestCreate(t *testing.T) {

	test_cases := []struct {
		name         string
		id           string
		channel      string
		expected_err error
	}{
		{"Invalid ID", "", "test", subscription.ErrInvalidID},
		{"Invalid Channel", "test", "", subscription.ErrInvalidChannel},
		{"Happy Path", "1234", "abcd", nil},
	}

	for _, tc := range test_cases {
		t.Run(tc.name, func(t *testing.T) {
			registry := &mock_registry{}
			service := sub_service.NewCreateSubscriptionService(registry, 1)
			result, err := service.Create(sub_service.CreateSubscriptionInput{
				ID:      tc.id,
				Channel: tc.channel,
			})

			if tc.expected_err != nil {

				if err == nil {
					t.Error("expected error, got nil")
				}

				if err != tc.expected_err {
					t.Errorf("expected %v, received %v", tc.expected_err, err)
				}

				if result != nil {
					t.Errorf("expect nil subscription, received %v", result)
				}

				if registry.called == true {
					t.Error("unexpected subscription registration")
				}
			} else {

				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if result == nil {
					t.Error("expected subscription, received nil")
				}

				if registry.called == false {
					t.Error("expected registry to have been called")
				}

				if registry.registered_sub == nil {
					t.Error("expected subscription to have been registered")
				}

				if registry.registered_sub != result {
					t.Errorf("expected Register() to have been called with subscription: %v, received %v", result, registry.registered_sub)
				}

			}

		})
	}

}
