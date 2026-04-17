package transport

import (
	"context"
	"net/url"

	"github.com/coder/websocket"
	"github.com/exanubes/appsync/internal/app"
)

func Dial(ctx context.Context, options app.DialOptions) (*Connection, error) {
	if options.Url == "" {
		return nil, app.ErrEmptyUrl
	}

	url, err := url.Parse(options.Url)
	if err != nil {
		return nil, err
	}

	conn, _, err := websocket.Dial(ctx, options.Url, &websocket.DialOptions{
		Host:         url.Hostname(),
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
func (conn *Connection) Close() error {
	return conn.ws.Close(websocket.StatusNormalClosure, websocket.StatusNormalClosure.String())
}
