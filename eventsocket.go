package eventsocket

import (
	"sync"
)

type Eventsocket struct {
	clientManager *clientManager
	roomManager   *roomManager
	eventManager  *eventManager
	mu            sync.Mutex
}

func New() *Eventsocket {
	em := newEventManager()

	es := &Eventsocket{
		clientManager: newClientManager(),
		roomManager: newRoomManager(&roomManagerConfig{
			eventManager: em,
		}),
		eventManager: em,
	}

	return es
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

	es.eventManager.triggerCreateClient(client)

	return client, nil
}

func (es *Eventsocket) RemoveClient(clientID string) {
	es.mu.Lock()

	client, exists := es.clientManager.getClient(clientID)
	if !exists {
		return
	}

	es.clientManager.removeClient(clientID)
	es.roomManager.disconnectClient(clientID)

	es.mu.Unlock()

	client.disconnect()

	es.eventManager.triggerRemoveClient(clientID)
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

func (es *Eventsocket) BroadcastToClient(clientID string, msg Message) error {
	return es.clientManager.broadcastToClient(clientID, msg)
}

func (es *Eventsocket) OnCreateClient(name string, handler CreateClientHandler) {
	es.eventManager.onCreateClient(name, handler)
}

func (es *Eventsocket) OnRemoveClient(name string, handler RemoveClientHandler) {
	es.eventManager.onRemoveClient(name, handler)
}

func (es *Eventsocket) OnJoinRoom(name string, handler JoinRoomHandler) {
	es.eventManager.onJoinRoom(name, handler)
}

func (es *Eventsocket) OnLeaveRoom(name string, handler LeaveRoomHandler) {
	es.eventManager.onLeaveRoom(name, handler)
}

func (es *Eventsocket) OnNewRoom(name string, handler NewRoomHandler) {
	es.eventManager.onNewRoom(name, handler)
}

func (es *Eventsocket) OnDeleteRoom(name string, handler DeleteRoomHandler) {
	es.eventManager.onDeleteRoom(name, handler)
}

func (es *Eventsocket) OffCreateClient(name string) {
	es.eventManager.offCreateClient(name)
}

func (es *Eventsocket) OffRemoveClient(name string) {
	es.eventManager.offRemoveClient(name)
}

func (es *Eventsocket) OffJoinRoom(name string) {
	es.eventManager.offJoinRoom(name)
}

func (es *Eventsocket) OffLeaveRoom(name string) {
	es.eventManager.offLeaveRoom(name)
}

func (es *Eventsocket) OffNewRoom(name string) {
	es.eventManager.offNewRoom(name)
}

func (es *Eventsocket) OffDeleteRoom(name string) {
	es.eventManager.offDeleteRoom(name)
}
