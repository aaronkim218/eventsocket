package hubsocket

import (
	"log/slog"

	"github.com/gofiber/contrib/websocket"
)

type ClientConfig[T any] struct {
	Metadata T
	Conn     *websocket.Conn
}

type Client[T any] struct {
	Metadata T
	conn     *websocket.Conn
	Write    chan WsMessage
	done     chan struct{}
}

func NewClient[T any](cfg *ClientConfig[T]) *Client[T] {
	c := &Client[T]{
		Metadata: cfg.Metadata,
		conn:     cfg.Conn,
		Write:    make(chan WsMessage),
		done:     make(chan struct{}),
	}

	go c.handleWriteClient()

	return c
}

func (c *Client[T]) Done() <-chan struct{} {
	return c.done
}

func (c *Client[T]) handleWriteClient() {
	for wsm := range c.Write {
		if err := c.conn.WriteJSON(wsm); err != nil {
			slog.Error("Error writing message to client. closing connection", slog.String("err", err.Error()))
			c.closeConn()
			return
		}
	}
}

func (c *Client[T]) closeConn() {
	if err := c.conn.Close(); err != nil {
		slog.Error("Error closing conn", slog.String("err", err.Error()))
	}
}
