package connection_test

import "github.com/exanubes/appsync/internal/app"

type noop_logger struct{}

func (n *noop_logger) Debug(_ string, _ ...any) {}
func (n *noop_logger) SetContext(_ string) app.Logger { return n }
