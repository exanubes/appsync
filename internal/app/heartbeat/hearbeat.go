package heartbeat

import (
	"context"
	"time"

	"github.com/exanubes/appsync/internal/app"
)

type Heartbeat struct {
	clock app.Clock
	reset chan struct{}
}

func New(clock app.Clock) *Heartbeat {
	return &Heartbeat{
		clock: clock,
		reset: make(chan struct{}, 1),
	}
}

func (heartbeat *Heartbeat) Start(ctx context.Context, timeout time.Duration) error {
	grace_period := timeout / 10
	heart_rate := timeout + grace_period

	timer := heartbeat.clock.NewTimer(heart_rate)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C():
			return app.ErrHeartbeatTimeout
		case <-heartbeat.reset:
			timer.Stop()
			timer.Reset(heart_rate)
		}
	}
}

func (heartbeat *Heartbeat) Reset() {
	if heartbeat == nil {
		return
	}

	select {
	case heartbeat.reset <- struct{}{}:
	default:
	}
}
