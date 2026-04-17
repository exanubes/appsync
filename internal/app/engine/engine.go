package engine

import (
	"context"
)

type Engine struct{}

func New() *Engine {
	return &Engine{}
}

func (engine *Engine) Start(ctx context.Context, input StartEngineInput) error {
	heartbeat := NewHeartbeat(input.Timeout)
	heartbeat.Start(ctx)
	return nil
}
