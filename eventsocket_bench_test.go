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
	select {
	case <-m.closed:
		return fmt.Errorf("connection closed")
	default:
		select {}
	}
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

	select {
	case <-m.closed:
	default:
		close(m.closed)
	}
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

func BenchmarkClientRemoval(b *testing.B) {
	es := eventsocket.New()

	clients := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		cfg := &eventsocket.CreateClientConfig{
			ID:   strconv.Itoa(i),
			Conn: newMockConn(),
		}
		client, err := es.CreateClient(cfg)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}
		clients[i] = client.ID()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		es.RemoveClient(clients[i])
	}
}

func BenchmarkMessageThroughput(b *testing.B) {
	es := eventsocket.New()

	cfg := &eventsocket.CreateClientConfig{
		ID:   "test-client",
		Conn: newMockConn(),
	}
	_, err := es.CreateClient(cfg)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	msg := eventsocket.Message{
		Type: "test",
		Data: json.RawMessage(`{"content":"benchmark message"}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := es.BroadcastToClient("test-client", msg)
		if err != nil {
			b.Fatalf("Failed to send message: %v", err)
		}
	}
}

func BenchmarkRoomJoinLeave(b *testing.B) {
	es := eventsocket.New()

	numClients := 1000
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
		roomID := strconv.Itoa(i % 10)

		err := es.AddClientToRoom(roomID, clientID)
		if err != nil {
			b.Fatalf("Failed to add client to room: %v", err)
		}

		es.RemoveClientFromRoom(roomID, clientID)
	}
}

func BenchmarkRoomBroadcast(b *testing.B) {
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

				err = es.AddClientToRoom(roomID, strconv.Itoa(i))
				if err != nil {
					b.Fatalf("Failed to add client to room: %v", err)
				}
			}

			msg := eventsocket.Message{
				Type: "room-broadcast",
				Data: json.RawMessage(`{"content":"hello room"}`),
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := es.BroadcastToRoom(roomID, msg)
				if err != nil {
					b.Fatalf("Failed to broadcast to room: %v", err)
				}
			}
		})
	}
}

func BenchmarkGlobalBroadcast(b *testing.B) {
	clientCounts := []int{100, 1000, 10000}

	for _, count := range clientCounts {
		b.Run(fmt.Sprintf("Clients-%d", count), func(b *testing.B) {
			es := eventsocket.New()

			for i := 0; i < count; i++ {
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
				Type: "global-broadcast",
				Data: json.RawMessage(`{"content":"hello everyone"}`),
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				es.BroadcastToAll(msg)
			}
		})
	}
}

func BenchmarkEventHandlerExecution(b *testing.B) {
	es := eventsocket.New()

	handlerCount := 10
	for i := 0; i < handlerCount; i++ {
		handlerName := fmt.Sprintf("handler-%d", i)
		es.OnCreateClient(handlerName, func(client *eventsocket.Client) {
			_ = client.ID()
		})
	}

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

func BenchmarkConcurrentClientOperations(b *testing.B) {
	es := eventsocket.New()
	var counter int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			clientID := strconv.Itoa(int(atomic.AddInt64(&counter, 1)))

			cfg := &eventsocket.CreateClientConfig{
				ID:   clientID,
				Conn: newMockConn(),
			}
			client, err := es.CreateClient(cfg)
			if err != nil {
				b.Fatalf("Failed to create client: %v", err)
			}

			msg := eventsocket.Message{
				Type: "test",
				Data: json.RawMessage(`{"content":"test"}`),
			}
			err = es.BroadcastToClient(client.ID(), msg)
			if err != nil {
				b.Fatalf("Failed to broadcast: %v", err)
			}

			es.RemoveClient(client.ID())
		}
	})
}

func BenchmarkConnectionChurn(b *testing.B) {
	es := eventsocket.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clientID := strconv.Itoa(i)
		cfg := &eventsocket.CreateClientConfig{
			ID:   clientID,
			Conn: newMockConn(),
		}

		client, err := es.CreateClient(cfg)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}

		es.RemoveClient(client.ID())
	}
}

func BenchmarkMessageBursts(b *testing.B) {
	es := eventsocket.New()

	numClients := 100
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

	msg := eventsocket.Message{
		Type: "burst",
		Data: json.RawMessage(`{"content":"burst message"}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			clientID := clientIDs[(i*10+j)%numClients]
			err := es.BroadcastToClient(clientID, msg)
			if err != nil {
				b.Fatalf("Failed to send burst message: %v", err)
			}
		}
	}
}

func BenchmarkLargeMessagePayloads(b *testing.B) {
	payloadSizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
	}

	for _, payload := range payloadSizes {
		b.Run(payload.name, func(b *testing.B) {
			es := eventsocket.New()

			cfg := &eventsocket.CreateClientConfig{
				ID:   "test-client",
				Conn: newMockConn(),
			}
			_, err := es.CreateClient(cfg)
			if err != nil {
				b.Fatalf("Failed to create client: %v", err)
			}

			data := make([]byte, payload.size)
			for i := range data {
				data[i] = byte('A' + (i % 26))
			}

			msg := eventsocket.Message{
				Type: "large",
				Data: json.RawMessage(data),
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := es.BroadcastToClient("test-client", msg)
				if err != nil {
					b.Fatalf("Failed to send large message: %v", err)
				}
			}
		})
	}
}

func BenchmarkRoomLifecycle(b *testing.B) {
	es := eventsocket.New()

	numClients := 50
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

	msg := eventsocket.Message{
		Type: "lifecycle",
		Data: json.RawMessage(`{"content":"room lifecycle test"}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		roomID := strconv.Itoa(i)

		err := es.AddClientToRoom(roomID, clientIDs[0])
		if err != nil {
			b.Fatalf("Failed to create room: %v", err)
		}

		for j := 1; j < 10; j++ {
			err := es.AddClientToRoom(roomID, clientIDs[j])
			if err != nil {
				b.Fatalf("Failed to populate room: %v", err)
			}
		}

		err = es.BroadcastToRoom(roomID, msg)
		if err != nil {
			b.Fatalf("Failed to broadcast to room: %v", err)
		}

		for j := 0; j < 10; j++ {
			es.RemoveClientFromRoom(roomID, clientIDs[j])
		}
	}
}

func BenchmarkRoomScalingPatterns(b *testing.B) {
	scenarios := []struct {
		name           string
		numRooms       int
		clientsPerRoom int
	}{
		{"ManySmallRooms", 1000, 5},
		{"FewLargeRooms", 10, 500},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			es := eventsocket.New()

			totalClients := scenario.numRooms * scenario.clientsPerRoom
			clientIDs := make([]string, totalClients)

			for i := 0; i < totalClients; i++ {
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

			clientIndex := 0
			for roomNum := 0; roomNum < scenario.numRooms; roomNum++ {
				roomID := strconv.Itoa(roomNum)
				for clientInRoom := 0; clientInRoom < scenario.clientsPerRoom; clientInRoom++ {
					err := es.AddClientToRoom(roomID, clientIDs[clientIndex])
					if err != nil {
						b.Fatalf("Failed to add client to room: %v", err)
					}
					clientIndex++
				}
			}

			msg := eventsocket.Message{
				Type: "scaling-test",
				Data: json.RawMessage(`{"content":"scaling pattern test"}`),
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				roomID := strconv.Itoa(i % scenario.numRooms)
				err := es.BroadcastToRoom(roomID, msg)
				if err != nil {
					b.Fatalf("Failed to broadcast to room: %v", err)
				}
			}
		})
	}
}

func BenchmarkMixedWorkload(b *testing.B) {
	es := eventsocket.New()

	numClients := 1000
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

	numRooms := 50
	for i := 0; i < numRooms; i++ {
		roomID := strconv.Itoa(i)
		for j := 0; j < 10; j++ {
			clientID := clientIDs[(i*10+j)%numClients]
			err := es.AddClientToRoom(roomID, clientID)
			if err != nil {
				b.Fatalf("Failed to add client to room: %v", err)
			}
		}
	}

	msg := eventsocket.Message{
		Type: "mixed-workload",
		Data: json.RawMessage(`{"content":"mixed workload test"}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		operation := i % 4
		switch operation {
		case 0:
			roomID := strconv.Itoa(i % numRooms)
			err := es.BroadcastToRoom(roomID, msg)
			if err != nil {
				b.Fatalf("Failed to broadcast to room: %v", err)
			}
		case 1:
			es.BroadcastToAll(msg)
		case 2:
			roomID := strconv.Itoa(i % numRooms)
			clientID := clientIDs[i%numClients]
			es.AddClientToRoom(roomID, clientID)
		case 3:
			roomID := strconv.Itoa(i % numRooms)
			clientID := clientIDs[i%numClients]
			es.RemoveClientFromRoom(roomID, clientID)
		}
	}
}
