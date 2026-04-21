package app

import "errors"

var ErrEmptyUrl = errors.New("Url is empty")
var ErrHandshakeTimeout = errors.New("Handshake timeout")
var ErrDuplicateMessage = errors.New("Message was already sent once")
var ErrSubscriptionInboxFull = errors.New("Subscription incoming message buffer exceeded")
