package engine

import (
	"context"

	"github.com/exanubes/appsync/internal/app"
)

type Engine struct{}

func New(hearbeat app.Heartbeat) *Engine {
	return &Engine{}
}

func (engine *Engine) Start(ctx context.Context) error {
	return nil
}
