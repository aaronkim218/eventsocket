package hubsocket

import (
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

type AddClientRequest[T any] struct {
	RoomId uuid.UUID
	Client *Client[T]
}

type Config[T any] struct {
	Workers         int
	Logger          *slog.Logger
	StatsInterval   time.Duration
	CleanupInterval time.Duration
	PluginRegistry  *PluginRegistry[T]
}

type Hub[T any] struct {
	activeRooms    map[uuid.UUID]*ActiveRoom[T]
	mu             sync.Mutex
	logger         *slog.Logger
	pluginRegistry *PluginRegistry[T]
}

func New[T any](cfg *Config[T]) *Hub[T] {
	hub := &Hub[T]{
		activeRooms:    make(map[uuid.UUID]*ActiveRoom[T]),
		logger:         cfg.Logger,
		pluginRegistry: cfg.PluginRegistry,
	}

	go hub.cleanup(cfg.CleanupInterval)
	go hub.stats(cfg.StatsInterval)

	return hub
}

func (h *Hub[T]) AddClient(client *Client[T], roomId uuid.UUID) {
	h.loadActiveRoom(roomId).handleClientJoin(client)
}

func (h *Hub[T]) loadActiveRoom(roomId uuid.UUID) *ActiveRoom[T] {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.activeRooms[roomId]
	if !ok {
		room = newActiveRoom[T](&activeRoomConfig[T]{
			roomId:         roomId,
			logger:         h.logger,
			pluginRegistry: h.pluginRegistry,
		})
		h.activeRooms[roomId] = room
	}

	return room
}

func (h *Hub[T]) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)

	for range ticker.C {
		h.mu.Lock()
		for roomId, room := range h.activeRooms {
			room.mu.RLock()
			if len(room.Clients) == 0 {
				delete(h.activeRooms, roomId)
				close(room.broadcast)
				h.logger.Info("Deleted room", slog.String("id", roomId.String()))
			}
			room.mu.RUnlock()
		}
		h.mu.Unlock()
	}
}

func (h *Hub[T]) stats(interval time.Duration) {
	ticker := time.NewTicker(interval)

	for range ticker.C {
		h.mu.Lock()
		h.logger.Info("Hub stats", slog.Int("num active rooms", len(h.activeRooms)))
		h.mu.Unlock()
	}
}
