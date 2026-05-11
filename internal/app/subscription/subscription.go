package subscription

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/exanubes/appsync/internal/app"
)

var ErrInvalidID = errors.New("Expected ID, received empty string")
var ErrInvalidChannel = errors.New("Expected channel, received empty string")

type Subscription struct {
	id      string
	channel string
	inbox   *inbox
	done    chan struct{}
	once    sync.Once
	active  bool
	mutex   sync.RWMutex
}

func New(sub_id string, channel string, buffer_size uint) (*Subscription, error) {
	if sub_id == "" {
		return nil, ErrInvalidID
	}

	if channel == "" {
		return nil, ErrInvalidChannel
	}

	done := make(chan struct{}, 1)
	return &Subscription{
		id:      sub_id,
		channel: channel,
		done:    done,
		inbox:   new_inbox(done, buffer_size),
		active:  true,
	}, nil

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

func (subscription *Subscription) Close() {
	subscription.once.Do(func() {
		close(subscription.done)
		subscription.mutex.Lock()
		subscription.active = false
		subscription.mutex.Unlock()
	})
}

func (subscription *Subscription) Active() bool {
	subscription.mutex.RLock()
	defer subscription.mutex.RUnlock()

	return subscription.active
}
