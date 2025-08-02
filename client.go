package eventsocket

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type ClientConfig struct {
	Conn        *websocket.Conn
	Eventsocket *Eventsocket
}

type Client struct {
	id              string
	conn            *websocket.Conn
	eventsocket     *Eventsocket
	mu              sync.RWMutex
	once            sync.Once
	disconnected    bool
	messageHandlers map[string]MessageHandler
	outgoing        chan Message
}

func NewClient(cfg *ClientConfig) *Client {
	c := &Client{
		conn:            cfg.Conn,
		eventsocket:     cfg.Eventsocket,
		messageHandlers: make(map[string]MessageHandler),
		outgoing:        make(chan Message),
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
		_ = c.conn.Close()
	}))
}

func (c *Client) read() {
	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			c.eventsocket.RemoveClient(c.id)
			return
		}

		c.handleMessage(msg)
	}
}

func (c *Client) write() {
	for msg := range c.outgoing {
		if err := c.conn.WriteJSON(msg); err != nil {
			c.eventsocket.RemoveClient(c.id)
		}
	}
}

func (c *Client) handleMessage(msg Message) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if handler, ok := c.messageHandlers[msg.Type]; ok {
		handler(msg.Data)
	}
}

func (c *Client) ID() string {
	return c.id
}
