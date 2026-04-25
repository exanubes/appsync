package runtime

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type Runtime struct {
	router    app.Router
	heartbeat app.Heartbeat
	inbox     app.Inbox
}

func New(inbox app.Inbox, router app.Router, heartbeat app.Heartbeat) *Runtime {
	return &Runtime{
		router:    router,
		heartbeat: heartbeat,
		inbox:     inbox,
	}
}
func (runtime *Runtime) Run(ctx context.Context) error {
	for {
		msg, err := runtime.inbox.Next(ctx)
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
