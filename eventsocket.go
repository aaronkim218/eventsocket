package eventsocket

import (
	"sync"
)

type Config struct {
	OnCreateClient CreateClientHandler
	OnRemoveClient RemoveClientHandler
	OnJoinRoom     JoinRoomHandler
	OnLeaveRoom    LeaveRoomHandler
	OnNewRoom      NewRoomHandler
	OnDeleteRoom   DeleteRoomHandler
}

type Eventsocket struct {
	clientManager  *clientManager
	roomManager    *roomManager
	mu             sync.Mutex
	onCreateClient CreateClientHandler
	onRemoveClient RemoveClientHandler
	onJoinRoom     JoinRoomHandler
	onLeaveRoom    LeaveRoomHandler
	onNewRoom      NewRoomHandler
	onDeleteRoom   DeleteRoomHandler
}

func New(cfg *Config) *Eventsocket {
	return &Eventsocket{
		clientManager:  newClientManager(),
		roomManager:    newRoomManager(),
		onCreateClient: cfg.OnCreateClient,
		onRemoveClient: cfg.OnRemoveClient,
		onJoinRoom:     cfg.OnJoinRoom,
		onLeaveRoom:    cfg.OnLeaveRoom,
		onNewRoom:      cfg.OnNewRoom,
		onDeleteRoom:   cfg.OnDeleteRoom,
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

	if es.onCreateClient != nil {
		es.onCreateClient(client)
	}

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
		if es.onDeleteRoom != nil {
			es.onDeleteRoom(roomID)
		}
	}

	client.disconnect()

	if es.onRemoveClient != nil {
		es.onRemoveClient(clientID)
	}
}

func (es *Eventsocket) AddClientToRoom(roomID string, clientID string) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	client, exists := es.clientManager.getClient(clientID)
	if !exists {
		return ErrClientNotFound
	}

	roomCreated := es.roomManager.addClientToRoom(roomID, client)

	if roomCreated && es.onNewRoom != nil {
		es.onNewRoom(roomID)
	}

	if es.onJoinRoom != nil {
		es.onJoinRoom(roomID, clientID)
	}

	return nil
}

func (es *Eventsocket) RemoveClientFromRoom(roomID string, clientID string) {
	roomDeleted := es.roomManager.removeClientFromRoom(roomID, clientID)

	if es.onLeaveRoom != nil {
		es.onLeaveRoom(roomID, clientID)
	}

	if roomDeleted && es.onDeleteRoom != nil {
		es.onDeleteRoom(roomID)
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
