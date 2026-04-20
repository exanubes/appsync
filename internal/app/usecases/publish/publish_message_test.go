package publish_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/usecases/publish"
)

type mock_authorizer struct {
	signature app.Signature
	err       error
}

func (mock *mock_authorizer) Authorize(_ context.Context, _ app.AuthorizeCommandInput) (app.Signature, error) {
	return mock.signature, mock.err
}

type mock_sender struct {
	err      error
	called   bool
	received app.Frame
}

func (mock *mock_sender) Send(_ context.Context, frame app.Frame) error {
	mock.called = true
	mock.received = frame
	return mock.err
}

type mock_frame struct{}

func (mock mock_frame) ID() string                   { return "test-id" }
func (mock mock_frame) Encode() (app.Payload, error) { return nil, nil }

type mockFrameBuilder struct {
	payload     app.Payload
	channel     string
	signature   app.Signature
	built_frame app.Frame
}

func (m *mockFrameBuilder) WithPayload(p app.Payload) app.FrameBuilder     { m.payload = p; return m }
func (m *mockFrameBuilder) WithChannel(c string) app.FrameBuilder          { m.channel = c; return m }
func (m *mockFrameBuilder) WithSignature(s app.Signature) app.FrameBuilder { m.signature = s; return m }
func (m *mockFrameBuilder) Build() app.Frame                               { return m.built_frame }

func TestPublish(t *testing.T) {
	auth_err := errors.New("auth failed")
	send_err := errors.New("send failed")
	signature := app.Signature{"Authorization": "sig-value"}
	payload := app.Payload("test-payload")
	destination := "test-channel"

	tests := []struct {
		name             string
		authorizer       *mock_authorizer
		sender           *mock_sender
		expect_err       error
		expect_send      bool
		expect_payload   app.Payload
		expect_channel   string
		expect_signature app.Signature
	}{
		{
			name:             "success",
			authorizer:       &mock_authorizer{signature: signature},
			sender:           &mock_sender{},
			expect_err:       nil,
			expect_send:      true,
			expect_payload:   payload,
			expect_channel:   destination,
			expect_signature: signature,
		},
		{
			name:        "authorizer error does not call send",
			authorizer:  &mock_authorizer{err: auth_err},
			sender:      &mock_sender{},
			expect_err:  auth_err,
			expect_send: false,
		},
		{
			name:             "writer error is returned",
			authorizer:       &mock_authorizer{signature: signature},
			sender:           &mock_sender{err: send_err},
			expect_err:       send_err,
			expect_send:      true,
			expect_payload:   payload,
			expect_channel:   destination,
			expect_signature: signature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := &mockFrameBuilder{built_frame: mock_frame{}}
			usecase := publish.NewPublishMessageUsecase(tt.authorizer, tt.sender)

			err := usecase.Publish(context.Background(), publish.PublishCommandInput{
				Destination: destination,
				Payload:     payload,
				Frame:       frame,
			})

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("got error %v, want %v", err, tt.expect_err)
			}

			if tt.sender.called != tt.expect_send {
				t.Errorf("sender.called = %v, want %v", tt.sender.called, tt.expect_send)
			}

			if tt.expect_send {
				if string(frame.payload) != string(tt.expect_payload) {
					t.Errorf("frame.payload = %q, want %q", frame.payload, tt.expect_payload)
				}
				if frame.channel != tt.expect_channel {
					t.Errorf("frame.channel = %q, want %q", frame.channel, tt.expect_channel)
				}
			}

			if tt.expect_signature != nil {
				for k, v := range tt.expect_signature {
					if frame.signature[k] != v {
						t.Errorf("frame.signature[%q] = %q, want %q", k, frame.signature[k], v)
					}
				}
			}
		})
	}
}
