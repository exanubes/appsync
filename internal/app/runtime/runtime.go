package runtime

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type Runtime struct {
	router    app.Router
	heartbeat app.Heartbeat
}

func New(router app.Router, heartbeat app.Heartbeat) *Runtime {
	return &Runtime{
		router:    router,
		heartbeat: heartbeat,
	}
}
func (runtime *Runtime) Run(ctx context.Context, inbox app.Inbox) error {
	for {
		msg, err := inbox.Next(ctx)
		if err != nil {
			return err
		}

		runtime.heartbeat.Reset()
		err = runtime.router.Handle(ctx, msg)

		if err != nil {
			return err
		}

	}
}
