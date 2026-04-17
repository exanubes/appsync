package engine

import (
	"time"

	"github.com/exanubes/appsync/internal/app/services/connection"
)

type StartEngineInput struct {
	Timeout    time.Duration
	Connection connection.Connection
}
