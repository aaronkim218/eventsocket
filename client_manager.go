package eventsocket

import "sync"

type clientManager struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func newClientManager() *clientManager {
	return &clientManager{
		clients: make(map[string]*Client),
	}
}

func (cm *clientManager) addClient(client *Client) {
	cm.mu.Lock()
	cm.clients[client.ID()] = client
	cm.mu.Unlock()
}

func (cm *clientManager) removeClient(clientID string) {
	cm.mu.Lock()
	delete(cm.clients, clientID)
	cm.mu.Unlock()
}

func (cm *clientManager) getClient(clientID string) (*Client, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.clients[clientID]
	return client, exists
}

func (cm *clientManager) broadcast(msg Message) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, client := range cm.clients {
		client.sendMessage(msg)
	}
}
