package publish_test

import (
	"context"
	"errors"
	"testing"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/protocol"
	"github.com/exanubes/appsync/internal/app/usecases/publish"
)

type mock_authorizer struct {
	signature app.Signature
	err       error
}

func (mock *mock_authorizer) Authorize(_ context.Context, _ app.AuthorizeCommandInput) (app.Signature, error) {
	return mock.signature, mock.err
}

type mock_batcher struct {
	result   []app.Batch
	called   bool
	received []app.Payload
}

func (mock *mock_batcher) Batch(events []app.Payload) []app.Batch {
	mock.called = true
	mock.received = events
	return mock.result
}

type mock_sender struct {
	errs     []error
	calls    int
	called   bool
	received app.Frame
}

func (mock *mock_sender) Send(_ context.Context, frame app.Frame) error {
	mock.called = true
	mock.received = frame
	var err error
	if mock.calls < len(mock.errs) {
		err = mock.errs[mock.calls]
	}
	mock.calls++
	return err
}

type mock_frame struct{}

func (mock mock_frame) ID() string                   { return "test-id" }
func (mock mock_frame) Encode() (app.Payload, error) { return nil, nil }

type mockFrameBuilder struct {
	batch       app.Batch
	payload     app.Payload
	channel     string
	signature   app.Signature
	built_frame app.Frame
}

func (m *mockFrameBuilder) WithPayload(p app.Payload) app.FrameBuilder     { m.payload = p; return m }
func (m *mockFrameBuilder) WithBatch(b app.Batch) app.FrameBuilder         { m.batch = b; return m }
func (m *mockFrameBuilder) WithChannel(c string) app.FrameBuilder          { m.channel = c; return m }
func (m *mockFrameBuilder) WithSignature(s app.Signature) app.FrameBuilder { m.signature = s; return m }
func (m *mockFrameBuilder) WithType(_ string) app.FrameBuilder             { return m }
func (m *mockFrameBuilder) WithID(_ string) app.FrameBuilder               { return m }
func (m *mockFrameBuilder) Build() app.Frame                               { return m.built_frame }

func TestPublish(t *testing.T) {
	auth_err := errors.New("auth failed")
	send_err := errors.New("send failed")
	signature := app.Signature{"Authorization": "sig-value"}
	payload := app.Payload("test-payload")
	destination := "test-channel"

	tests := []struct {
		name                string
		authorizer          app.RequestAuthorizer
		override_authorizer *mock_authorizer
		sender              *mock_sender
		batcher             *mock_batcher
		expect_err          error
		expect_send_calls   int
		expect_success      bool
		expect_failures     int
		expect_payload      app.Payload
		expect_channel      string
		expect_signature    app.Signature
	}{
		{
			name:              "success",
			authorizer:        &mock_authorizer{signature: signature},
			sender:            &mock_sender{},
			batcher:           &mock_batcher{result: []app.Batch{{payload}}},
			expect_err:        nil,
			expect_send_calls: 1,
			expect_success:    true,
			expect_failures:   0,
			expect_payload:    payload,
			expect_channel:    destination,
			expect_signature:  signature,
		},
		{
			name:              "authorizer error does not call send",
			authorizer:        &mock_authorizer{err: auth_err},
			sender:            &mock_sender{},
			batcher:           &mock_batcher{result: []app.Batch{{payload}}},
			expect_err:        auth_err,
			expect_send_calls: 0,
		},
		{
			name:                "override authorizer is used instead of default",
			authorizer:          &mock_authorizer{err: auth_err},
			override_authorizer: &mock_authorizer{signature: app.Signature{"Authorization": "override-sig"}},
			sender:              &mock_sender{},
			batcher:             &mock_batcher{result: []app.Batch{{payload}}},
			expect_err:          nil,
			expect_send_calls:   1,
			expect_success:      true,
			expect_failures:     0,
			expect_payload:      payload,
			expect_channel:      destination,
			expect_signature:    app.Signature{"Authorization": "override-sig"},
		},
		{
			name:              "writer error is returned",
			authorizer:        &mock_authorizer{signature: signature},
			sender:            &mock_sender{errs: []error{send_err}},
			batcher:           &mock_batcher{result: []app.Batch{{payload}}},
			expect_err:        send_err,
			expect_send_calls: 1,
			expect_payload:    payload,
			expect_channel:    destination,
			expect_signature:  signature,
		},
		{
			name:              "nil authorizer returns error",
			authorizer:        nil,
			sender:            &mock_sender{},
			batcher:           &mock_batcher{},
			expect_err:        app.ErrPublishAuthorizerMissing,
			expect_send_calls: 0,
		},
		{
			name:                "override authorizer used when default is nil",
			authorizer:          nil,
			override_authorizer: &mock_authorizer{signature: signature},
			sender:              &mock_sender{},
			batcher:             &mock_batcher{result: []app.Batch{{payload}}},
			expect_err:          nil,
			expect_send_calls:   1,
			expect_success:      true,
			expect_failures:     0,
			expect_payload:      payload,
			expect_channel:      destination,
			expect_signature:    signature,
		},
		{
			name:       "batch publish error populates failures",
			authorizer: &mock_authorizer{signature: signature},
			sender: &mock_sender{errs: []error{
				protocol.BatchPublishError{Failures: []protocol.FailedEvent{{Index: 0}}},
			}},
			batcher:           &mock_batcher{result: []app.Batch{{payload}}},
			expect_err:        nil,
			expect_send_calls: 1,
			expect_success:    false,
			expect_failures:   1,
		},
		{
			name:              "multiple batches all succeed",
			authorizer:        &mock_authorizer{signature: signature},
			sender:            &mock_sender{},
			batcher:           &mock_batcher{result: []app.Batch{{payload}, {payload}}},
			expect_err:        nil,
			expect_send_calls: 2,
			expect_success:    true,
			expect_failures:   0,
		},
		{
			name:       "batch publish errors accumulate across batches",
			authorizer: &mock_authorizer{signature: signature},
			sender: &mock_sender{errs: []error{
				protocol.BatchPublishError{Failures: []protocol.FailedEvent{{Index: 0}}},
				protocol.BatchPublishError{Failures: []protocol.FailedEvent{{Index: 0}}},
			}},
			batcher:           &mock_batcher{result: []app.Batch{{payload}, {payload}}},
			expect_err:        nil,
			expect_send_calls: 2,
			expect_success:    false,
			expect_failures:   2,
		},
		{
			name:              "non-batch error aborts remaining batches",
			authorizer:        &mock_authorizer{signature: signature},
			sender:            &mock_sender{errs: []error{send_err}},
			batcher:           &mock_batcher{result: []app.Batch{{payload}, {payload}}},
			expect_err:        send_err,
			expect_send_calls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := &mockFrameBuilder{built_frame: mock_frame{}}
			usecase := publish.NewPublishMessageUsecase(tt.authorizer, tt.sender, tt.batcher)

			var override_authorizer app.RequestAuthorizer
			if tt.override_authorizer != nil {
				override_authorizer = tt.override_authorizer
			}
			output, err := usecase.Publish(context.Background(), publish.PublishCommandInput{
				Destination: destination,
				Events:      []app.Payload{payload},
				Frame:       frame,
				Authorizer:  override_authorizer,
			})

			if !errors.Is(err, tt.expect_err) {
				t.Errorf("got error %v, want %v", err, tt.expect_err)
			}

			if tt.sender.calls != tt.expect_send_calls {
				t.Errorf("sender.calls = %d, want %d", tt.sender.calls, tt.expect_send_calls)
			}

			if tt.expect_payload != nil {
				if len(frame.batch) == 0 || string(frame.batch[0]) != string(tt.expect_payload) {
					t.Errorf("frame.batch[0] = %q, want %q", frame.batch[0], tt.expect_payload)
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

			if output != nil {
				if output.Success != tt.expect_success {
					t.Errorf("Success = %v, want %v", output.Success, tt.expect_success)
				}
				if len(output.Failures) != tt.expect_failures {
					t.Errorf("len(Failures) = %d, want %d", len(output.Failures), tt.expect_failures)
				}
			}
		})
	}
}
