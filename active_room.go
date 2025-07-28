package hubsocket

import (
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

type activeRoomConfig[T any] struct {
	roomId         uuid.UUID
	logger         *slog.Logger
	pluginRegistry *PluginRegistry[T]
}

type ActiveRoom[T any] struct {
	RoomId         uuid.UUID
	Clients        map[*Client[T]]struct{}
	Mu             sync.RWMutex
	broadcast      chan BroadcastMessage[T]
	logger         *slog.Logger
	pluginRegistry *PluginRegistry[T]
}

func newActiveRoom[T any](cfg *activeRoomConfig[T]) *ActiveRoom[T] {
	ar := &ActiveRoom[T]{
		RoomId:         cfg.roomId,
		Clients:        make(map[*Client[T]]struct{}),
		broadcast:      make(chan BroadcastMessage[T]),
		logger:         cfg.logger,
		pluginRegistry: cfg.pluginRegistry,
	}

	go ar.handleBroadcast()

	return ar
}

func (ar *ActiveRoom[T]) handleBroadcast() {
	for bm := range ar.broadcast {
		ar.handleBroadcastMessage(bm)
	}
}

func (ar *ActiveRoom[T]) handleReadClient(client *Client[T]) {
	for {
		var wsm WsMessage
		if err := client.conn.ReadJSON(&wsm); err != nil {
			ar.logger.Error("Error reading message from client. closing connection", slog.String("error", err.Error()))
			client.closeConn()
			ar.handleClientLeave(client)
			return
		}

		ar.broadcast <- BroadcastMessage[T]{
			Client:    client,
			WsMessage: wsm,
		}
	}
}

func (ar *ActiveRoom[T]) handleBroadcastMessage(msg BroadcastMessage[T]) {
	if err := ar.pluginRegistry.handleBroadcastMessage(ar, msg); err != nil {
		ar.logger.Error("Plugin failed to handle message",
			slog.String("err", err.Error()),
			slog.String("type", string(msg.WsMessage.Type)),
			slog.Any("payload", msg.WsMessage.Payload),
		)
	}
}

func (ar *ActiveRoom[T]) handleClientJoin(client *Client[T]) {
	ar.Mu.Lock()
	ar.Clients[client] = struct{}{}
	ar.Mu.Unlock()
	go ar.handleReadClient(client)

	if err := ar.pluginRegistry.handleClientJoin(ar, client); err != nil {
		ar.logger.Error("Plugin failed to handle message", slog.String("err", err.Error()))
	}
}

func (ar *ActiveRoom[T]) handleClientLeave(client *Client[T]) {
	ar.Mu.Lock()
	delete(ar.Clients, client)
	ar.Mu.Unlock()

	if err := ar.pluginRegistry.handleClientLeave(ar, client); err != nil {
		ar.logger.Error("Plugin failed to handle message", slog.String("err", err.Error()))
	}

	close(client.Write)
	client.done <- struct{}{}
}
