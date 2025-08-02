package eventsocket

import (
	"sync"
)

type roomManager struct {
	rooms map[string]*room
	mu    sync.RWMutex
}

func newRoomManager() *roomManager {
	return &roomManager{
		rooms: make(map[string]*room),
	}
}

func (rm *roomManager) addClientToRoom(roomID string, client *Client) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		room = newRoom()
		rm.rooms[roomID] = room
	}

	room.addClient(client)
}

func (rm *roomManager) removeClientFromRoom(roomID string, clientID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return
	}

	room.removeClient(clientID)
	if len(room.getClients()) == 0 {
		delete(rm.rooms, roomID)
	}
}

// TODO: optimize
func (rm *roomManager) disconnectClient(clientID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for id, room := range rm.rooms {
		room.removeClient(clientID)
		if len(room.getClients()) == 0 {
			delete(rm.rooms, id)
		}
	}
}

func (rm *roomManager) broadcastToRoom(roomID string, msg Message) error {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return ErrRoomNotFound
	}

	room.broadcast(msg)
	return nil
}

func (rm *roomManager) broadcastToRoomExcept(roomID string, clientID string, msg Message) error {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return ErrRoomNotFound
	}

	return room.broadcastExcept(clientID, msg)
}
