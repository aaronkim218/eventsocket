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
	if room.size() == 0 {
		delete(rm.rooms, roomID)
		return true
	}

	return false
}

// returns 2 slices. first indicates ids of rooms that were deleted. second indicates rooms that client left as a result of disconnecting
func (rm *roomManager) disconnectClient(clientID string) ([]string, []string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	deletedRooms := make([]string, 0)
	roomsLeft := make([]string, 0)
	for id, room := range rm.rooms {
		wasMember := room.removeClient(clientID)
		if wasMember {
			roomsLeft = append(roomsLeft, id)
		}

		if room.size() == 0 {
			delete(rm.rooms, id)
			deletedRooms = append(deletedRooms, id)
		}
	}

	return deletedRooms, roomsLeft
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
