package subscription

import (
	"context"
	"encoding/json"

	"github.com/exanubes/appsync/internal/app"
)

type Subscription struct {
	id      string
	channel string
	inbox   *inbox
}

func New(sub_id string, channel string, buffer_size uint) *Subscription {
	return &Subscription{
		id:      sub_id,
		channel: channel,
		inbox:   new_inbox(buffer_size),
	}
}

func (subscription *Subscription) Deliver(ctx context.Context, payload app.Payload) error {
	return subscription.inbox.Enqueue(ctx, payload)
}

func (subscription *Subscription) Next(ctx context.Context) ([]byte, error) {
	return subscription.inbox.Next(ctx)
}

func (subscription *Subscription) Decode(ctx context.Context, value any) error {
	payload, err := subscription.inbox.Next(ctx)

	if err != nil {
		return err
	}

	return json.Unmarshal(payload, value)
}
