package eventsocket

import (
	"sync"
)

type Eventsocket struct {
	clientManager        *clientManager
	roomManager          *roomManager
	mu                   sync.RWMutex
	createClientHandlers map[string]CreateClientHandler
	removeClientHandlers map[string]RemoveClientHandler
	joinRoomHandlers     map[string]JoinRoomHandler
	leaveRoomHandlers    map[string]LeaveRoomHandler
	newRoomHandlers      map[string]NewRoomHandler
	deleteRoomHandlers   map[string]DeleteRoomHandler
}

func New() *Eventsocket {
	return &Eventsocket{
		clientManager:        newClientManager(),
		roomManager:          newRoomManager(),
		createClientHandlers: make(map[string]CreateClientHandler),
		removeClientHandlers: make(map[string]RemoveClientHandler),
		joinRoomHandlers:     make(map[string]JoinRoomHandler),
		leaveRoomHandlers:    make(map[string]LeaveRoomHandler),
		newRoomHandlers:      make(map[string]NewRoomHandler),
		deleteRoomHandlers:   make(map[string]DeleteRoomHandler),
	}
}

type CreateClientConfig struct {
	ID   string
	Conn Conn
}

func (es *Eventsocket) CreateClient(cfg *CreateClientConfig) (*Client, error) {
	if _, exists := es.clientManager.getClient(cfg.ID); exists {
		return nil, ErrClientExists
	}

	clientCfg := &clientConfig{
		ID:          cfg.ID,
		Conn:        cfg.Conn,
		Eventsocket: es,
	}

	client := newClient(clientCfg)
	es.clientManager.addClient(client)

	es.triggerCreateClient(client)

	return client, nil
}

func (es *Eventsocket) RemoveClient(clientID string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	client, exists := es.clientManager.getClient(clientID)
	if !exists {
		return
	}

	es.clientManager.removeClient(clientID)
	deletedRoomIDs := es.roomManager.disconnectClient(clientID)

	for _, roomID := range deletedRoomIDs {
		es.triggerDeleteRoom(roomID)
	}

	client.disconnect()

	es.triggerRemoveClient(clientID)
}

func (es *Eventsocket) AddClientToRoom(roomID string, clientID string) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	client, exists := es.clientManager.getClient(clientID)
	if !exists {
		return ErrClientNotFound
	}

	roomCreated := es.roomManager.addClientToRoom(roomID, client)

	if roomCreated {
		es.triggerNewRoom(roomID)
	}

	es.triggerJoinRoom(roomID, clientID)

	return nil
}

func (es *Eventsocket) RemoveClientFromRoom(roomID string, clientID string) {
	roomDeleted := es.roomManager.removeClientFromRoom(roomID, clientID)

	es.triggerLeaveRoom(roomID, clientID)

	if roomDeleted {
		es.triggerDeleteRoom(roomID)
	}
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

func (es *Eventsocket) BroadcastToClient(clientID string, msg Message) error {
	return es.clientManager.broadcastToClient(clientID, msg)
}

func (es *Eventsocket) OnCreateClient(name string, handler CreateClientHandler) {
	if handler == nil {
		return
	}

	es.mu.Lock()
	defer es.mu.Unlock()

	es.createClientHandlers[name] = handler
}

func (es *Eventsocket) OnRemoveClient(name string, handler RemoveClientHandler) {
	if handler == nil {
		return
	}

	es.mu.Lock()
	defer es.mu.Unlock()

	es.removeClientHandlers[name] = handler
}

func (es *Eventsocket) OnJoinRoom(name string, handler JoinRoomHandler) {
	if handler == nil {
		return
	}

	es.mu.Lock()
	defer es.mu.Unlock()

	es.joinRoomHandlers[name] = handler
}

func (es *Eventsocket) OnLeaveRoom(name string, handler LeaveRoomHandler) {
	if handler == nil {
		return
	}

	es.mu.Lock()
	defer es.mu.Unlock()

	es.leaveRoomHandlers[name] = handler
}

func (es *Eventsocket) OnNewRoom(name string, handler NewRoomHandler) {
	if handler == nil {
		return
	}

	es.mu.Lock()
	defer es.mu.Unlock()

	es.newRoomHandlers[name] = handler
}

func (es *Eventsocket) OnDeleteRoom(name string, handler DeleteRoomHandler) {
	if handler == nil {
		return
	}

	es.mu.Lock()
	defer es.mu.Unlock()

	es.deleteRoomHandlers[name] = handler
}

func (es *Eventsocket) OffCreateClient(name string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	delete(es.createClientHandlers, name)
}

func (es *Eventsocket) OffRemoveClient(name string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	delete(es.removeClientHandlers, name)
}

func (es *Eventsocket) OffJoinRoom(name string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	delete(es.joinRoomHandlers, name)
}

func (es *Eventsocket) OffLeaveRoom(name string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	delete(es.leaveRoomHandlers, name)
}

func (es *Eventsocket) OffNewRoom(name string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	delete(es.newRoomHandlers, name)
}

func (es *Eventsocket) OffDeleteRoom(name string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	delete(es.deleteRoomHandlers, name)
}

func (es *Eventsocket) triggerCreateClient(client *Client) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, handler := range es.createClientHandlers {
		handler(client)
	}
}

func (es *Eventsocket) triggerRemoveClient(clientID string) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, handler := range es.removeClientHandlers {
		handler(clientID)
	}
}

func (es *Eventsocket) triggerJoinRoom(roomID, clientID string) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, handler := range es.joinRoomHandlers {
		handler(roomID, clientID)
	}
}

func (es *Eventsocket) triggerLeaveRoom(roomID, clientID string) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, handler := range es.leaveRoomHandlers {
		handler(roomID, clientID)
	}
}

func (es *Eventsocket) triggerNewRoom(roomID string) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, handler := range es.newRoomHandlers {
		handler(roomID)
	}
}

func (es *Eventsocket) triggerDeleteRoom(roomID string) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, handler := range es.deleteRoomHandlers {
		handler(roomID)
	}
}
