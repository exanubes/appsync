package engine

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/exanubes/appsync/internal/app"
)

type Engine struct {
	logger      app.Logger
	io          IO
	runtime     Runtime
	err_channel chan error
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func New(runtime Runtime, io IO, logger app.Logger) *Engine {
	return &Engine{
		io:          io,
		logger:      logger.SetContext("Engine"),
		runtime:     runtime,
		err_channel: make(chan error, 3),
	}
}

func (engine *Engine) Start(ctx context.Context, input StartEngineInput) {
	engine.ctx, engine.cancel = context.WithCancel(ctx)
	engine.wg.Add(3)
	go func() {
		engine.err_channel <- engine.io.Read(engine.ctx, input.Ingress)
		engine.logger.Debug("Exitted ingress loop")
		engine.wg.Done()
		engine.cancel()
	}()

	go func() {
		engine.err_channel <- engine.io.Write(engine.ctx, input.Egress)

		engine.logger.Debug("Exitted egress loop")
		engine.wg.Done()
		engine.cancel()
	}()

	go func() {
		engine.err_channel <- engine.runtime.Run(engine.ctx, input.Ingress)
		engine.logger.Debug("Exitted runtime loop")
		engine.wg.Done()
		engine.cancel()
	}()
}

func (engine *Engine) Close(ctx context.Context) error {
	engine.cancel()
	engine.wg.Wait()
	engine.logger.Debug("Waitgroup done")
	index := 3
	var error error
	for index > 0 {
		index -= 1

		select {
		case <-ctx.Done():
			return ctx.Err()

		case err := <-engine.err_channel:
			if err != nil && errors.Is(err, context.Canceled) {
				error = fmt.Errorf("%w\n%w", error, err)
			}
		default:
		}
	}

	engine.logger.Debug("I/O Engine shutdown complete")
	return error
}
