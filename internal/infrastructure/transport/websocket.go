package transport

import (
	"context"

	"github.com/coder/websocket"
	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/services/connection"
)

type Websocket struct{}

func New() *Websocket {
	return &Websocket{}
}

func (*Websocket) Dial(ctx context.Context, options connection.DialOptions) (connection.Connection, error) {
	if options.Url == nil {
		return nil, app.ErrEmptyUrl
	}

	conn, _, err := websocket.Dial(ctx, options.Url.String(), &websocket.DialOptions{
		Host:         options.Url.Hostname(),
		Subprotocols: options.Subprotocols,
	})

	if err != nil {
		if conn != nil {
			conn.Close(websocket.StatusProtocolError, "Failed to connect")
		}

		return nil, err
	}

	return &Connection{
		ws: conn,
	}, nil
}

type Connection struct {
	ws *websocket.Conn
}

func (conn *Connection) Read(ctx context.Context) ([]byte, error) {
	_, data, err := conn.ws.Read(ctx)
	return data, err
}
func (conn *Connection) Write(ctx context.Context, data []byte) error {
	return conn.ws.Write(ctx, websocket.MessageText, data)
}
func (conn *Connection) Close(ctx context.Context) error {
	return conn.ws.Close(websocket.StatusNormalClosure, websocket.StatusNormalClosure.String())
}
