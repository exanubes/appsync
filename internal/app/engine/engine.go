package engine

import (
	"context"
	"errors"
	"sync"

	"github.com/exanubes/appsync/internal/app"
)

type Engine struct {
	logger      app.Logger
	io          IO
	runtime     Runtime
	heartbeat   app.Heartbeat
	err_channel chan error
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	state       ConnectionState
}

var managed_goroutines_count = 4

func New(heartbeat app.Heartbeat, runtime Runtime, io IO, state ConnectionState, logger app.Logger) *Engine {
	return &Engine{
		io:          io,
		logger:      logger.SetContext("Engine"),
		runtime:     runtime,
		heartbeat:   heartbeat,
		state:       state,
		err_channel: make(chan error, managed_goroutines_count),
	}
}

func (engine *Engine) Start(ctx context.Context, input StartEngineInput) {
	engine.ctx, engine.cancel = context.WithCancel(ctx)
	engine.wg.Add(managed_goroutines_count)
	go func() {
		err := engine.io.Read(engine.ctx)
		engine.state.Close(err)
		engine.err_channel <- err
		engine.logger.Debug("Exitted ingress loop")
		engine.wg.Done()
		engine.cancel()
	}()

	go func() {
		err := engine.io.Write(engine.ctx)
		engine.state.Close(err)
		engine.err_channel <- err
		engine.logger.Debug("Exitted egress loop")
		engine.wg.Done()
		engine.cancel()
	}()

	go func() {
		err := engine.runtime.Run(engine.ctx)
		engine.state.Close(err)
		engine.err_channel <- err
		engine.logger.Debug("Exitted runtime loop")
		engine.wg.Done()
		engine.cancel()
	}()

	go func() {
		err := engine.heartbeat.Start(engine.ctx, input.Timeout)
		engine.state.Close(err)
		engine.err_channel <- err
		engine.logger.Debug("Exitted heartbeat loop")
		engine.wg.Done()
		engine.cancel()
	}()
}

func (engine *Engine) Close(ctx context.Context) error {
	engine.cancel()
	engine.wg.Wait()
	index := managed_goroutines_count
	var error error
	for index > 0 {
		index -= 1

		select {
		case <-ctx.Done():
			return ctx.Err()

		case err := <-engine.err_channel:
			if err != nil && !errors.Is(err, context.Canceled) {
				error = errors.Join(err)
			}
		default:
		}
	}

	engine.logger.Debug("I/O Engine shutdown complete")
	return error
}
