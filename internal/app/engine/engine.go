package engine

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type Engine struct {
	logger app.Logger
}

func New(logger app.Logger) *Engine {
	return &Engine{
		logger: logger.SetContext("Engine"),
	}
}

func (engine *Engine) Start(ctx context.Context, input StartEngineInput) error {
	NewHeartbeat(input.Timeout)
	engine.logger.Debug("Engine.Start")
	return input.Connection.Close()
}
