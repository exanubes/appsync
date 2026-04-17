package heartbeat

import (
	"context"
	"time"
)

type Heartbeat struct {
	timeout      time.Duration
	grace_period time.Duration
	reset        chan struct{}
}

func New(timeout time.Duration) *Heartbeat {
	return &Heartbeat{
		timeout:      timeout,
		grace_period: timeout / 10,
		reset:        make(chan struct{}, 1),
	}
}

func (heartbeat *Heartbeat) Start(ctx context.Context) <-chan error {
	return make(chan error, 1)
}

func (heartbeat *Heartbeat) Reset() {}
