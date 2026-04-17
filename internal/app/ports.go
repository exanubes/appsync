package app

import (
	"context"
)

type DialOptions struct {
	Url          string
	Subprotocols []string
}

type Reader interface {
	Read(context.Context) ([]byte, error)
}

type Writer interface {
	Write(context.Context, []byte) error
}

type Closer interface {
	Close() error
}

type Connection interface {
	Reader
	Writer
	Closer
}
