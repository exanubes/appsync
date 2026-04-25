package subscribe_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
	sub_service "github.com/exanubes/appsync/internal/app/services/subscription"
	"github.com/exanubes/appsync/internal/app/subscription"
	"github.com/exanubes/appsync/internal/app/usecases/subscribe"
)

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

type mock_frame struct {
	id string
}

func (m mock_frame) ID() string                   { return m.id }
func (m mock_frame) Encode() (app.Payload, error) { return nil, nil }

type mock_frame_builder struct {
	frame_type  string
	channel     string
	signature   app.Signature
	built_frame app.Frame
}

func (m *mock_frame_builder) WithPayload(_ app.Payload) app.FrameBuilder { return m }
func (m *mock_frame_builder) WithChannel(c string) app.FrameBuilder      { m.channel = c; return m }
func (m *mock_frame_builder) WithSignature(s app.Signature) app.FrameBuilder {
	m.signature = s
	return m
}
func (m *mock_frame_builder) WithType(t string) app.FrameBuilder { m.frame_type = t; return m }
func (m *mock_frame_builder) WithID(_ string) app.FrameBuilder   { return m }
func (m *mock_frame_builder) Build() app.Frame                   { return m.built_frame }

type mock_create_subscription struct {
	sub   *subscription.Subscription
	input sub_service.CreateSubscriptionInput
}

func (m *mock_create_subscription) Create(input sub_service.CreateSubscriptionInput) (*subscription.Subscription, error) {
	m.input = input
	return m.sub, nil
}

func TestSubscribeChannel(t *testing.T) {
	auth_err := errors.New("auth failed")
	send_err := errors.New("send failed")
	signature := app.Signature{"Authorization": "sig-value"}
	channel := "test-channel"
	frame_id := "test-id"
	sub, _ := subscription.New(frame_id, channel, 1)

	tests := []struct {
		name           string
		authorizer     *mock_authorizer
		sender         *mock_sender
		create_sub     *mock_create_subscription
		expect_err     error
		expect_sub_id  string
		expect_sub     *subscription.Subscription
		expect_type    string
		expect_channel string
		expect_sig     app.Signature
	}{
		{
			name:           "success",
			authorizer:     &mock_authorizer{signature: signature},
			sender:         &mock_sender{},
			create_sub:     &mock_create_subscription{sub: sub},
			expect_err:     nil,
			expect_sub_id:  frame_id,
			expect_sub:     sub,
			expect_type:    protocol.TypeSubscribe,
			expect_channel: channel,
			expect_sig:     signature,
		},
		{
			name:       "authorize error",
			authorizer: &mock_authorizer{err: auth_err},
			sender:     &mock_sender{},
			create_sub: &mock_create_subscription{},
			expect_err: auth_err,
		},
		{
			name:       "send error",
			authorizer: &mock_authorizer{signature: signature},
			sender:     &mock_sender{err: send_err},
			create_sub: &mock_create_subscription{},
			expect_err: send_err,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := &mock_frame_builder{built_frame: mock_frame{id: frame_id}}
			usecase := subscribe.NewSubscribeChannelUsecase(tt.authorizer, tt.sender, tt.create_sub)

			output, err := usecase.Execute(context.Background(), subscribe.SubscribeCommandInput{
				Channel: channel,
				Frame:   frame,
			})

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("got error %v, want %v", err, tt.expect_err)
			}

			if tt.expect_err != nil {
				if output != nil {
					t.Errorf("expected nil output on error, got %v", output)
				}
				return
			}

			if output.SubID != tt.expect_sub_id {
				t.Errorf("output.SubID = %q, want %q", output.SubID, tt.expect_sub_id)
			}

			if output.Subscription != tt.expect_sub {
				t.Errorf("output.Subscription = %v, want %v", output.Subscription, tt.expect_sub)
			}

			if frame.frame_type != tt.expect_type {
				t.Errorf("frame.frame_type = %q, want %q", frame.frame_type, tt.expect_type)
			}

			if frame.channel != tt.expect_channel {
				t.Errorf("frame.channel = %q, want %q", frame.channel, tt.expect_channel)
			}

			for k, v := range tt.expect_sig {
				if frame.signature[k] != v {
					t.Errorf("frame.signature[%q] = %q, want %q", k, frame.signature[k], v)
				}
			}

			if tt.create_sub.input.ID != frame_id {
				t.Errorf("create_sub.input.ID = %q, want %q", tt.create_sub.input.ID, frame_id)
			}

			if tt.create_sub.input.Channel != channel {
				t.Errorf("create_sub.input.Channel = %q, want %q", tt.create_sub.input.Channel, channel)
			}
		})
	}
}
