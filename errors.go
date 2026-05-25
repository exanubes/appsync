package appsync

import "github.com/exanubes/appsync/internal/app"

var (
	ErrEmptyUrl              = app.ErrEmptyUrl
	ErrHandshakeTimeout      = app.ErrHandshakeTimeout
	ErrDuplicateMessage      = app.ErrDuplicateMessage
	ErrSubscriptionInboxFull = app.ErrSubscriptionInboxFull
	ErrSubscriptionClosed    = app.ErrSubscriptionClosed
	ErrSubscriptionNotFound  = app.ErrSubscriptionNotFound
	ErrHeartbeatTimeout      = app.ErrHeartbeatTimeout
	ErrConnectionClosed      = app.ErrConnectionClosed
)
