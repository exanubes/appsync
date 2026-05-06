package heartbeat_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/heartbeat"
)

type fake_timer struct {
	mu         sync.Mutex
	c          chan time.Time
	stop_count int
	reset_dur  time.Duration
}

func (f *fake_timer) C() <-chan time.Time {
	return f.c
}

func (f *fake_timer) Stop() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stop_count++
	return true
}

func (f *fake_timer) Reset(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reset_dur = d
}

func (f *fake_timer) get_stop_count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stop_count
}

func (f *fake_timer) get_reset_dur() time.Duration {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.reset_dur
}

type fake_clock struct {
	timer *fake_timer
}

func (f *fake_clock) NewTimer(_ time.Duration) app.Timer {
	return f.timer
}

func (f *fake_clock) Now() time.Time {
	return time.Time{}
}

type start_result struct {
	err error
}

func run_start(hb *heartbeat.Heartbeat, ctx context.Context, timeout time.Duration) <-chan start_result {
	ch := make(chan start_result, 1)
	go func() {
		ch <- start_result{hb.Start(ctx, timeout)}
	}()
	return ch
}

func TestStart(t *testing.T) {
	const timeout = 100 * time.Millisecond

	tests := []struct {
		name       string
		trigger    func(timer *fake_timer, cancel context.CancelFunc)
		expect_err error
	}{
		{
			name: "context cancellation returns ctx.Err",
			trigger: func(_ *fake_timer, cancel context.CancelFunc) {
				cancel()
			},
			expect_err: context.Canceled,
		},
		{
			name: "timer fires returns ErrHeartbeatTimeout",
			trigger: func(timer *fake_timer, _ context.CancelFunc) {
				timer.c <- time.Now()
			},
			expect_err: app.ErrHeartbeatTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timer := &fake_timer{c: make(chan time.Time, 1)}
			clock := &fake_clock{timer: timer}
			hb := heartbeat.New(clock)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ch := run_start(hb, ctx, timeout)

			tt.trigger(timer, cancel)

			select {
			case result := <-ch:
				if !errors.Is(result.err, tt.expect_err) {
					t.Errorf("Start() error = %v, want %v", result.err, tt.expect_err)
				}
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Start did not return in time")
			}
		})
	}

	t.Run("Reset calls timer.Stop and timer.Reset with heart_rate then continues", func(t *testing.T) {
		const timeout = 200 * time.Millisecond
		expected_heart_rate := timeout + timeout/10

		timer := &fake_timer{c: make(chan time.Time, 1)}
		clock := &fake_clock{timer: timer}
		hb := heartbeat.New(clock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ch := run_start(hb, ctx, timeout)

		hb.Reset()
		// give the select loop time to drain the reset channel before asserting side effects
		time.Sleep(10 * time.Millisecond)

		if timer.get_stop_count() < 1 {
			t.Errorf("timer.stop_count = %d, want >= 1", timer.get_stop_count())
		}
		if timer.get_reset_dur() != expected_heart_rate {
			t.Errorf("timer.reset_dur = %v, want %v", timer.get_reset_dur(), expected_heart_rate)
		}

		select {
		case result := <-ch:
			t.Fatalf("Start returned early with error %v", result.err)
		default:
		}

		timer.c <- time.Now()
		select {
		case result := <-ch:
			if !errors.Is(result.err, app.ErrHeartbeatTimeout) {
				t.Errorf("Start() error = %v, want %v", result.err, app.ErrHeartbeatTimeout)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Start did not return after timer fired")
		}
	})
}

func TestReset(t *testing.T) {
	t.Run("nil receiver does not panic", func(t *testing.T) {
		var hb *heartbeat.Heartbeat
		hb.Reset()
	})

	t.Run("second consecutive Reset does not block", func(t *testing.T) {
		timer := &fake_timer{c: make(chan time.Time)}
		clock := &fake_clock{timer: timer}
		hb := heartbeat.New(clock)

		done := make(chan struct{}, 1)
		go func() {
			hb.Reset()
			hb.Reset()
			done <- struct{}{}
		}()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Reset blocked on second call")
		}
	})
}
