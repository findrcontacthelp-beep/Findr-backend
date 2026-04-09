package ws

import (
	"sync"

	"go.uber.org/zap"
)

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[string]*Client
	log   *zap.Logger
}

func NewHub(log *zap.Logger) *Hub {
	return &Hub{
		rooms: make(map[string]map[string]*Client),
		log:   log,
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[client.ChatID]; !ok {
		h.rooms[client.ChatID] = make(map[string]*Client)
	}
	h.rooms[client.ChatID][client.UserUUID] = client
	h.log.Info("client registered", zap.String("chatID", client.ChatID), zap.String("userUUID", client.UserUUID))
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[client.ChatID]; ok {
		if _, exists := room[client.UserUUID]; exists {
			delete(room, client.UserUUID)
			close(client.Send)
			if len(room) == 0 {
				delete(h.rooms, client.ChatID)
			}
		}
	}
	h.log.Info("client unregistered", zap.String("chatID", client.ChatID), zap.String("userUUID", client.UserUUID))
}

func (h *Hub) BroadcastToChat(chatID string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, ok := h.rooms[chatID]; ok {
		for _, client := range room {
			select {
			case client.Send <- data:
			default:
				h.log.Warn("client send buffer full, dropping message", zap.String("userUUID", client.UserUUID))
			}
		}
	}
}

func (h *Hub) IsUserOnline(chatID, userUUID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, ok := h.rooms[chatID]; ok {
		_, exists := room[userUUID]
		return exists
	}
	return false
}

func (h *Hub) GetOnlineUsers(chatID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var users []string
	if room, ok := h.rooms[chatID]; ok {
		for uuidVal := range room {
			users = append(users, uuidVal)
		}
	}
	return users
}
