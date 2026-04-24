package clock

import "time"

type Timer struct {
	current *time.Timer
}

func (timer *Timer) C() <-chan time.Time {
	return timer.current.C
}

func (timer *Timer) Stop() bool {
	return timer.current.Stop()
}

func (timer *Timer) Reset(duration time.Duration) {
	timer.current.Reset(duration)
}
