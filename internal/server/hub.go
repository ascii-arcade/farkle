package server

import (
	"encoding/json"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/ascii-arcade/farkle/internal/lobby"
	"github.com/ascii-arcade/farkle/internal/player"
)

type hub struct {
	clients    map[*client]bool
	broadcast  chan Message
	register   chan *client
	unregister chan *client

	lobbies map[string]*lobby.Lobby

	logger *slog.Logger
	mu     sync.Mutex
}

func newHub(logger *slog.Logger) *hub {
	h := &hub{
		clients:    make(map[*client]bool),
		broadcast:  make(chan Message),
		logger:     logger,
		register:   make(chan *client),
		unregister: make(chan *client),

		lobbies: make(map[string]*lobby.Lobby),
	}

	return h
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.logger.Info("registering client", "client", c)
			h.addClient(c)
		case c := <-h.unregister:
			h.logger.Info("unregistering client", "client", c)
			h.removeClient(c)
		}
	}
}

func (h *hub) monitorBroadcast() {
	for msg := range h.broadcast {
		h.broadcastMessage(msg)
	}
}

func (h *hub) addClient(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = true
}

func (h *hub) removeClient(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		c.active = false
		delete(h.clients, c)
	}
}

func (h *hub) broadcastMessage(msg Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		if msg.from == c {
			continue
		}

		if _, err := c.conn.Write(msg.toBytes()); err != nil {
			logger.Error("Failed to send message", "error", err)
		}
	}
}

func (h *hub) createLobby(host *player.Player, lobbyName string) *lobby.Lobby {
	lobby := lobby.NewLobby(lobbyName, host)
	h.lobbies[lobby.Id] = lobby
	h.addLobby(lobby)
	return lobby
}

func (h *hub) addLobby(lobby *lobby.Lobby) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lobbies[lobby.Id] = lobby
}

func (h *hub) startTimedBroadcast() {
	for {
		h.mu.Lock()
		lobbies := make([]*lobby.Lobby, 0, len(h.lobbies))
		for _, lobby := range h.lobbies {
			lobbies = append(lobbies, lobby)
		}
		slices.SortFunc(lobbies, func(a, b *lobby.Lobby) int {
			if a.CreatedAt.Before(b.CreatedAt) {
				return -1
			}
			if a.CreatedAt.After(b.CreatedAt) {
				return 1
			}
			return 0
		})
		b, err := json.Marshal(lobbies)
		if err != nil {
			h.logger.Error("Failed to marshal lobby", "error", err)
			continue
		}
		msg := Message{
			Channel: ChannelLobby,
			Type:    MessageTypeList,
			Data:    b,
		}
		h.mu.Unlock()

		h.broadcast <- msg

		time.Sleep(time.Second)
	}
}

func (h *hub) getLobby(id string) *lobby.Lobby {
	h.mu.Lock()
	defer h.mu.Unlock()
	lobby, ok := h.lobbies[id]
	if !ok {
		return nil
	}
	return lobby
}
