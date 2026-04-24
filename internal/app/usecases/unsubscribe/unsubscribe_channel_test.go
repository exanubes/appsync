package unsubscribe_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
	"github.com/exanubes/appsync/internal/app/subscription"
	"github.com/exanubes/appsync/internal/app/usecases/unsubscribe"
)

type mock_registry struct {
	sub     *subscription.Subscription
	removed string
}

func (m *mock_registry) Get(_ string) *subscription.Subscription { return m.sub }
func (m *mock_registry) Remove(id string)                        { m.removed = id }

type mock_authorizer struct {
	signature app.Signature
	err       error
}

func (m *mock_authorizer) Authorize(_ context.Context, _ app.AuthorizeCommandInput) (app.Signature, error) {
	return m.signature, m.err
}

type mock_sender struct {
	err    error
	called bool
}

func (m *mock_sender) Send(_ context.Context, _ app.Frame) error {
	m.called = true
	return m.err
}

type mock_frame struct{}

func (m mock_frame) ID() string                   { return "test-id" }
func (m mock_frame) Encode() (app.Payload, error) { return nil, nil }

type mock_frame_builder struct {
	frame_type string
	signature  app.Signature
	id         string
}

func (m *mock_frame_builder) WithPayload(_ app.Payload) app.FrameBuilder      { return m }
func (m *mock_frame_builder) WithChannel(_ string) app.FrameBuilder           { return m }
func (m *mock_frame_builder) WithSignature(s app.Signature) app.FrameBuilder  { m.signature = s; return m }
func (m *mock_frame_builder) WithType(t string) app.FrameBuilder              { m.frame_type = t; return m }
func (m *mock_frame_builder) WithID(id string) app.FrameBuilder               { m.id = id; return m }
func (m *mock_frame_builder) Build() app.Frame                                { return mock_frame{} }

func active_sub() *subscription.Subscription {
	return subscription.New("sub-id", "test-channel", 0)
}

func inactive_sub() *subscription.Subscription {
	sub := subscription.New("sub-id", "test-channel", 0)
	sub.Close()
	return sub
}

func TestUnsubscribeChannel(t *testing.T) {
	const subscriptionId = "sub-id"

	auth_err := errors.New("auth failed")
	send_err := errors.New("send failed")
	signature := app.Signature{"Authorization": "sig-value"}

	tests := []struct {
		name             string
		registry         *mock_registry
		authorizer       *mock_authorizer
		sender           *mock_sender
		expect_err       error
		expect_send      bool
		expect_sub_closed bool
		expect_removed   string
	}{
		{
			name:       "subscription not found",
			registry:   &mock_registry{sub: nil},
			authorizer: &mock_authorizer{},
			sender:     &mock_sender{},
			expect_err: app.ErrSubscriptionClosed,
		},
		{
			name:       "inactive subscription",
			registry:   &mock_registry{sub: inactive_sub()},
			authorizer: &mock_authorizer{},
			sender:     &mock_sender{},
			expect_err: app.ErrSubscriptionClosed,
		},
		{
			name:       "authorize error stops execution",
			registry:   &mock_registry{sub: active_sub()},
			authorizer: &mock_authorizer{err: auth_err},
			sender:     &mock_sender{},
			expect_err: auth_err,
		},
		{
			name:        "send error does not clean up",
			registry:    &mock_registry{sub: active_sub()},
			authorizer:  &mock_authorizer{signature: signature},
			sender:      &mock_sender{err: send_err},
			expect_err:  send_err,
			expect_send: true,
		},
		{
			name:              "success",
			registry:          &mock_registry{sub: active_sub()},
			authorizer:        &mock_authorizer{signature: signature},
			sender:            &mock_sender{},
			expect_err:        nil,
			expect_send:       true,
			expect_sub_closed: true,
			expect_removed:    subscriptionId,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := &mock_frame_builder{}
			usecase := unsubscribe.NewUnsubscribeChannelUsecase(tt.registry, tt.authorizer, tt.sender)

			err := usecase.Execute(context.Background(), unsubscribe.UnsubscribeChannelCommandInput{
				SubscriptionId: subscriptionId,
				Frame:          frame,
			})

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("got error %v, want %v", err, tt.expect_err)
			}

			if tt.sender.called != tt.expect_send {
				t.Errorf("sender.called = %v, want %v", tt.sender.called, tt.expect_send)
			}

			if tt.expect_send {
				if frame.frame_type != protocol.TypeUnsubscribe {
					t.Errorf("frame.frame_type = %q, want %q", frame.frame_type, protocol.TypeUnsubscribe)
				}
				if frame.id != subscriptionId {
					t.Errorf("frame.id = %q, want %q", frame.id, subscriptionId)
				}
				for k, v := range tt.authorizer.signature {
					if frame.signature[k] != v {
						t.Errorf("frame.signature[%q] = %q, want %q", k, frame.signature[k], v)
					}
				}
			}

			if tt.expect_sub_closed && tt.registry.sub != nil {
				if tt.registry.sub.Active() {
					t.Error("subscription should be closed after successful unsubscribe")
				}
			}

			if !tt.expect_sub_closed && tt.registry.sub != nil && tt.expect_err != app.ErrSubscriptionClosed {
				if !tt.registry.sub.Active() {
					t.Error("subscription should remain active when execution fails")
				}
			}

			if tt.registry.removed != tt.expect_removed {
				t.Errorf("registry.removed = %q, want %q", tt.registry.removed, tt.expect_removed)
			}
		})
	}
}
