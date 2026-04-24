package clock

import (
	"time"

	"github.com/exanubes/appsync/internal/app"
)

type Clock struct{}

func New() *Clock {
	return &Clock{}
}

func (clock *Clock) NewTimer(duration time.Duration) app.Timer {
	return &Timer{current: time.NewTimer(duration)}
}
