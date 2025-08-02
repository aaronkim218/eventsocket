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

// returns true if room was created
func (rm *roomManager) addClientToRoom(roomID string, client *Client) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		room = newRoom()
		rm.rooms[roomID] = room
	}

	room.addClient(client)

	return !exists
}

// returns true if room was deleted
func (rm *roomManager) removeClientFromRoom(roomID string, clientID string) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return false
	}

	room.removeClient(clientID)
	if len(room.getClients()) == 0 {
		delete(rm.rooms, roomID)
		return true
	}

	return false
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
