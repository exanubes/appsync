package runtime

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/queue"
)

type Runtime struct {
	router app.Router
}

func New(router app.Router) *Runtime {
	return &Runtime{
		router: router,
	}
}
func (runtime *Runtime) Run(ctx context.Context, inbox *queue.IngressQueue) error {
	for {
		msg, err := inbox.Next(ctx)
		if err != nil {
			return err
		}

		err = runtime.router.Handle(ctx, msg)

		if err != nil {
			return err
		}

	}
}
