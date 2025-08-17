package eventsocket

import "sync"

type eventManager struct {
	createClientHandlers map[string]CreateClientHandler
	removeClientHandlers map[string]RemoveClientHandler
	joinRoomHandlers     map[string]JoinRoomHandler
	leaveRoomHandlers    map[string]LeaveRoomHandler
	newRoomHandlers      map[string]NewRoomHandler
	deleteRoomHandlers   map[string]DeleteRoomHandler
	mu                   sync.RWMutex
}

func newEventManager() *eventManager {
	return &eventManager{
		createClientHandlers: make(map[string]CreateClientHandler),
		removeClientHandlers: make(map[string]RemoveClientHandler),
		joinRoomHandlers:     make(map[string]JoinRoomHandler),
		leaveRoomHandlers:    make(map[string]LeaveRoomHandler),
		newRoomHandlers:      make(map[string]NewRoomHandler),
		deleteRoomHandlers:   make(map[string]DeleteRoomHandler),
	}
}

func (em *eventManager) onCreateClient(name string, handler CreateClientHandler) {
	if handler == nil {
		return
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	em.createClientHandlers[name] = handler
}

func (em *eventManager) onRemoveClient(name string, handler RemoveClientHandler) {
	if handler == nil {
		return
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	em.removeClientHandlers[name] = handler
}

func (em *eventManager) onJoinRoom(name string, handler JoinRoomHandler) {
	if handler == nil {
		return
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	em.joinRoomHandlers[name] = handler
}

func (em *eventManager) onLeaveRoom(name string, handler LeaveRoomHandler) {
	if handler == nil {
		return
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	em.leaveRoomHandlers[name] = handler
}

func (em *eventManager) onNewRoom(name string, handler NewRoomHandler) {
	if handler == nil {
		return
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	em.newRoomHandlers[name] = handler
}

func (em *eventManager) onDeleteRoom(name string, handler DeleteRoomHandler) {
	if handler == nil {
		return
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	em.deleteRoomHandlers[name] = handler
}

func (em *eventManager) offCreateClient(name string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	delete(em.createClientHandlers, name)
}

func (em *eventManager) offRemoveClient(name string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	delete(em.removeClientHandlers, name)
}

func (em *eventManager) offJoinRoom(name string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	delete(em.joinRoomHandlers, name)
}

func (em *eventManager) offLeaveRoom(name string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	delete(em.leaveRoomHandlers, name)
}

func (em *eventManager) offNewRoom(name string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	delete(em.newRoomHandlers, name)
}

func (em *eventManager) offDeleteRoom(name string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	delete(em.deleteRoomHandlers, name)
}

func (em *eventManager) triggerCreateClient(client *Client) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	for _, handler := range em.createClientHandlers {
		go handler(client)
	}
}

func (em *eventManager) triggerRemoveClient(clientID string) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	for _, handler := range em.removeClientHandlers {
		go handler(clientID)
	}
}

func (em *eventManager) triggerJoinRoom(roomID, clientID string) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	for _, handler := range em.joinRoomHandlers {
		go handler(roomID, clientID)
	}
}

func (em *eventManager) triggerLeaveRoom(roomID, clientID string) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	for _, handler := range em.leaveRoomHandlers {
		go handler(roomID, clientID)
	}
}

func (em *eventManager) triggerNewRoom(roomID string) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	for _, handler := range em.newRoomHandlers {
		go handler(roomID)
	}
}

func (em *eventManager) triggerDeleteRoom(roomID string) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	for _, handler := range em.deleteRoomHandlers {
		go handler(roomID)
	}
}
