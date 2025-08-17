package eventsocket

import (
	"sync"
)

type roomManager struct {
	rooms        map[string]*room
	eventManager *eventManager
	mu           sync.RWMutex
}

type roomManagerConfig struct {
	eventManager *eventManager
}

func newRoomManager(cfg *roomManagerConfig) *roomManager {
	return &roomManager{
		rooms:        make(map[string]*room),
		eventManager: cfg.eventManager,
	}
}

func (rm *roomManager) addClientToRoom(roomID string, client *Client) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		room = newRoom()
		rm.rooms[roomID] = room
		rm.eventManager.triggerNewRoom(roomID)
	}

	room.addClient(client)
	rm.eventManager.triggerJoinRoom(roomID, client.ID())
}

func (rm *roomManager) removeClientFromRoom(roomID string, clientID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return
	}

	room.removeClient(clientID)
	rm.eventManager.triggerRemoveClient(clientID)

	if room.size() == 0 {
		delete(rm.rooms, roomID)
		rm.eventManager.triggerDeleteRoom(roomID)
	}
}

func (rm *roomManager) disconnectClient(clientID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for id, room := range rm.rooms {
		removed := room.removeClient(clientID)
		if removed {
			rm.eventManager.triggerLeaveRoom(id, clientID)
		}

		if room.size() == 0 {
			delete(rm.rooms, id)
			rm.eventManager.triggerDeleteRoom(id)
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
