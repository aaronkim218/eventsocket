package eventsocket

import (
	"errors"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

var ErrNoHandler = errors.New("no handler for message type")

type ClientConfig struct {
	Conn        *websocket.Conn
	Eventsocket *Eventsocket
	OnError     func(error)
}

type Client struct {
	// client ID
	id string

	// connection
	conn *websocket.Conn

	// event socket
	eventsocket *Eventsocket

	// state management
	mu           sync.RWMutex
	once         sync.Once
	disconnected bool

	// message handling
	messageHandlers map[string]MessageHandler
	outgoing        chan Message

	// event handlers
	onError func(error)
}

func NewClient(cfg *ClientConfig) *Client {
	c := &Client{
		conn:            cfg.Conn,
		eventsocket:     cfg.Eventsocket,
		messageHandlers: make(map[string]MessageHandler),
		outgoing:        make(chan Message),
		onError:         cfg.OnError,
	}

	go c.read()
	go c.write()

	return c
}

func (c *Client) OnMessage(messageType string, handler MessageHandler) {
	c.mu.Lock()
	c.messageHandlers[messageType] = handler
	c.mu.Unlock()
}

func (c *Client) RemoveMessageHandler(messageType string) {
	c.mu.Lock()
	delete(c.messageHandlers, messageType)
	c.mu.Unlock()
}

func (c *Client) sendMessage(msg Message) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.disconnected {
		c.outgoing <- msg
	}
}

func (c *Client) disconnect() {
	c.once.Do((func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.disconnected = true
		close(c.outgoing)
		if err := c.conn.Close(); err != nil {
			c.onError(err)
		}
	}))
}

func (c *Client) read() {
	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			// TODO: what to do about this error
			c.onError(err)
			c.eventsocket.RemoveClient(c.id)
			return
		}

		if err := c.handleMessage(msg); err != nil {
			// TODO: what to do about this error
			c.onError(err)
		}
	}
}

func (c *Client) write() {
	for msg := range c.outgoing {
		if err := c.conn.WriteJSON(msg); err != nil {
			// TODO: what to do about this error
			c.onError(err)
			c.eventsocket.RemoveClient(c.id)
		}
	}
}

func (c *Client) handleMessage(msg Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	handler, ok := c.messageHandlers[msg.Type]
	if !ok {
		return ErrNoHandler
	}

	return handler(msg.Data)
}

func (c *Client) ID() string {
	return c.id
}
