package eventsocket

import (
	"sync"
)

type Eventsocket struct {
	clientManager *clientManager
	roomManager   *roomManager
	mu            sync.Mutex
}

func New() *Eventsocket {
	return &Eventsocket{
		clientManager: newClientManager(),
		roomManager:   newRoomManager(),
	}
}

type CreateClientConfig struct {
	Conn Conn
}

func (es *Eventsocket) CreateClient(cfg *CreateClientConfig) *Client {
	clientCfg := &ClientConfig{
		Conn:        cfg.Conn,
		Eventsocket: es,
	}

	client := NewClient(clientCfg)
	es.clientManager.addClient(client)
	return client
}

func (es *Eventsocket) RemoveClient(clientID string) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	client, exists := es.clientManager.getClient(clientID)
	if !exists {
		return ErrClientNotFound
	}

	es.clientManager.removeClient(clientID)
	es.roomManager.disconnectClient(clientID)
	client.disconnect()
	return nil
}

func (es *Eventsocket) AddClientToRoom(roomID string, clientID string) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	client, exists := es.clientManager.getClient(clientID)
	if !exists {
		return ErrClientNotFound
	}

	es.roomManager.addClientToRoom(roomID, client)

	return nil
}

func (es *Eventsocket) RemoveClientFromRoom(roomID string, clientID string) {
	es.roomManager.removeClientFromRoom(roomID, clientID)
}

func (es *Eventsocket) BroadcastToAll(msg Message) {
	es.clientManager.broadcast(msg)
}

func (es *Eventsocket) BroadcastToRoom(roomID string, msg Message) error {
	return es.roomManager.broadcastToRoom(roomID, msg)
}

func (es *Eventsocket) BroadcastToRoomExcept(roomID string, clientID string, msg Message) error {
	return es.roomManager.broadcastToRoomExcept(roomID, clientID, msg)
}
