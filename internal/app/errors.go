package app

import "errors"

var ErrEmptyUrl = errors.New("Url is empty")
var ErrHandshakeTimeout = errors.New("Handshake timeout")
var ErrDuplicateMessage = errors.New("Message was already sent once")
var ErrSubscriptionInboxFull = errors.New("Subscription incoming message buffer exceeded")
var ErrSubscriptionClosed = errors.New("Subscription is closed")
var ErrSubscriptionNotFound = errors.New("Subscription not found")
var ErrHeartbeatTimeout = errors.New("Heartbeat timeout")
