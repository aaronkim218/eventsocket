package eventsocket_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/aaronkim218/eventsocket"
)

type mockConn struct {
	closed chan struct{}
	mu     sync.Mutex
}

func newMockConn() *mockConn {
	return &mockConn{
		closed: make(chan struct{}),
	}
}

func (m *mockConn) ReadJSON(v any) error {
	<-m.closed
	return fmt.Errorf("connection closed")
}

func (m *mockConn) WriteJSON(v any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-m.closed:
		return fmt.Errorf("connection closed")
	default:
		return nil
	}
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	close(m.closed)
	return nil
}

func BenchmarkClientCreation(b *testing.B) {
	es := eventsocket.New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := &eventsocket.CreateClientConfig{
			ID:   strconv.Itoa(i),
			Conn: newMockConn(),
		}
		_, err := es.CreateClient(cfg)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}
	}
}

func BenchmarkClientCreationParallel(b *testing.B) {
	es := eventsocket.New()
	var counter int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cfg := &eventsocket.CreateClientConfig{
				ID:   strconv.Itoa(int(atomic.AddInt64(&counter, 1))),
				Conn: newMockConn(),
			}
			_, err := es.CreateClient(cfg)
			if err != nil {
				b.Fatalf("Failed to create client: %v", err)
			}
		}
	})
}

func BenchmarkClientCreationWithMessageHandlers(b *testing.B) {
	es := eventsocket.New()

	handler1 := func(data json.RawMessage) {}
	handler2 := func(data json.RawMessage) {}
	handler3 := func(data json.RawMessage) {}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := &eventsocket.CreateClientConfig{
			ID:   strconv.Itoa(i),
			Conn: newMockConn(),
		}
		client, err := es.CreateClient(cfg)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}

		client.OnMessage("type1", handler1)
		client.OnMessage("type2", handler2)
		client.OnMessage("type3", handler3)
	}
}

func BenchmarkClientRemoval(b *testing.B) {
	es := eventsocket.New()

	clients := make([]*eventsocket.Client, b.N)
	for i := 0; i < b.N; i++ {
		cfg := &eventsocket.CreateClientConfig{
			ID:   strconv.Itoa(i),
			Conn: newMockConn(),
		}
		client, err := es.CreateClient(cfg)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}
		clients[i] = client
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		es.RemoveClient(clients[i].ID())
	}
}

func BenchmarkBroadcastToAll(b *testing.B) {
	es := eventsocket.New()

	numClients := 1000
	for i := 0; i < numClients; i++ {
		cfg := &eventsocket.CreateClientConfig{
			ID:   strconv.Itoa(i),
			Conn: newMockConn(),
		}
		_, err := es.CreateClient(cfg)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}
	}

	msg := eventsocket.Message{
		Type: "broadcast",
		Data: json.RawMessage(`{"content":"hello everyone"}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		es.BroadcastToAll(msg)
	}
}

func BenchmarkBroadcastToRoom(b *testing.B) {
	roomSizes := []int{10, 100, 1000}
	for _, size := range roomSizes {
		b.Run(fmt.Sprintf("RoomSize-%d", size), func(b *testing.B) {
			es := eventsocket.New()
			roomID := "test-room"

			for i := 0; i < size; i++ {
				cfg := &eventsocket.CreateClientConfig{
					ID:   strconv.Itoa(i),
					Conn: newMockConn(),
				}
				_, err := es.CreateClient(cfg)
				if err != nil {
					b.Fatalf("Failed to create client: %v", err)
				}
				es.AddClientToRoom(roomID, strconv.Itoa(i))
			}

			msg := eventsocket.Message{
				Type: "room-broadcast",
				Data: json.RawMessage(`{"content":"hello room"}`),
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				es.BroadcastToRoom(roomID, msg)
			}
		})
	}
}

func BenchmarkRoomOperationsAtScale(b *testing.B) {
	es := eventsocket.New()

	numClients := 10000
	clientIDs := make([]string, numClients)
	for i := 0; i < numClients; i++ {
		clientID := strconv.Itoa(i)
		clientIDs[i] = clientID
		cfg := &eventsocket.CreateClientConfig{
			ID:   clientID,
			Conn: newMockConn(),
		}
		_, err := es.CreateClient(cfg)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clientID := clientIDs[i%numClients]
		roomID := strconv.Itoa(i % 100)
		es.AddClientToRoom(roomID, clientID)
		es.RemoveClientFromRoom(roomID, clientID)
	}
}
