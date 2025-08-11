package eventsocket

import (
	"sync"
)

type room struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func newRoom() *room {
	return &room{
		clients: make(map[string]*Client),
	}
}

func (r *room) addClient(client *Client) {
	r.mu.Lock()
	r.clients[client.ID()] = client
	r.mu.Unlock()
}

// returns true if client was a member and removed from the room
func (r *room) removeClient(clientID string) bool {
	r.mu.Lock()
	_, ok := r.clients[clientID]
	delete(r.clients, clientID)
	r.mu.Unlock()
	return ok
}

func (r *room) size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

func (r *room) broadcast(msg Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, client := range r.clients {
		client.sendMessage(msg)
	}
}

func (r *room) broadcastExcept(clientID string, msg Message) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.clients[clientID]
	if !exists {
		return ErrClientNotFound
	}

	for id, c := range r.clients {
		if id != clientID {
			c.sendMessage(msg)
		}
	}

	return nil
}
