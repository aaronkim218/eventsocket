package eventsocket

import "sync"

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

func (es *Eventsocket) AddClient(client *Client) {
	es.clientManager.addClient(client)
}

func (es *Eventsocket) RemoveClient(clientID string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	client, exists := es.clientManager.getClient(clientID)
	if !exists {
		// TODO: return error?
		return
	}

	es.clientManager.removeClient(clientID)
	es.roomManager.disconnectClient(clientID)
	client.disconnect()
}

func (es *Eventsocket) AddClientToRoom(roomID string, clientID string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	client, exists := es.clientManager.getClient(clientID)
	if !exists {
		return // TODO: return error?
	}

	es.roomManager.addClientToRoom(roomID, client)
}

func (es *Eventsocket) RemoveClientFromRoom(roomID string, clientID string) {
	es.roomManager.removeClientFromRoom(roomID, clientID)
}

func (es *Eventsocket) BroadcastToAll(msg Message) {
	es.clientManager.broadcast(msg)
}

func (es *Eventsocket) BroadcastToRoom(roomID string, msg Message) {
	es.roomManager.broadcastToRoom(roomID, msg)
}

func (es *Eventsocket) BroadcastToRoomExcept(roomID string, clientID string, msg Message) error {
	return es.roomManager.broadcastToRoomExcept(roomID, clientID, msg)
}
