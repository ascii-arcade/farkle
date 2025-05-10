package server

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/ascii-arcade/farkle/internal/lobby"
	"github.com/ascii-arcade/farkle/internal/message"
	"github.com/ascii-arcade/farkle/internal/player"
)

type hub struct {
	players    map[*player.Player]bool
	broadcast  chan message.Message
	register   chan *player.Player
	unregister chan *player.Player

	lobbies map[string]*lobby.Lobby

	logger *slog.Logger
	mu     sync.Mutex
}

func newHub(logger *slog.Logger) *hub {
	h := &hub{
		players:    make(map[*player.Player]bool),
		broadcast:  make(chan message.Message),
		logger:     logger,
		register:   make(chan *player.Player),
		unregister: make(chan *player.Player),

		lobbies: make(map[string]*lobby.Lobby),
	}

	return h
}

func (h *hub) run() {
	for {
		select {
		case p := <-h.register:
			h.logger.Info("registering client", "player", p.Id)
			go h.handleMessages(p)
			h.addPlayer(p)
		case c := <-h.unregister:
			h.logger.Info("unregistering client", "client", c)
			h.removePlayer(c)
		}
	}
}

func (h *hub) monitorBroadcast() {
	for msg := range h.broadcast {
		h.broadcastMessage(msg)
	}
}

func (h *hub) addPlayer(p *player.Player) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.players[p] = true
}

func (h *hub) removePlayer(p *player.Player) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.players[p]; ok {
		for _, lobby := range h.lobbies {
			lobby.RemovePlayer(p)
			if len(lobby.Players) == 0 {
				delete(h.lobbies, lobby.Code)
			}
		}
		_ = p.Close()
		delete(h.players, p)
	}
}

func (h *hub) broadcastMessage(msg message.Message, players ...*player.Player) {
	if h.mu.TryLock() {
		defer h.mu.Unlock()
	}

	if len(players) > 0 {
		for _, p := range players {
			if p == nil {
				continue
			}
			if err := p.SendMessage(msg); err != nil {
				h.logger.Error("Failed to send message", "error", err)
			}
		}
		return
	}

	for p := range h.players {
		if p == nil {
			continue
		}

		if msg.IsFromPlayer(p.Id) {
			continue
		}

		if err := p.SendMessage(msg); err != nil {
			logger.Error("Failed to send message", "error", err)
		}
	}
}

func (h *hub) createLobby(host *player.Player) *lobby.Lobby {
	lobby := lobby.NewLobby(host)
	h.addLobby(lobby)
	return lobby
}

func (h *hub) addLobby(lobby *lobby.Lobby) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lobbies[lobby.Code] = lobby
}

func (h *hub) getLobby(code string) *lobby.Lobby {
	h.mu.Lock()
	defer h.mu.Unlock()
	lobby, ok := h.lobbies[strings.ToUpper(code)]
	if !ok {
		return nil
	}
	return lobby
}

func (h *hub) removePlayerFromLobby(player *player.Player) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, lobby := range h.lobbies {
		lobby.RemovePlayer(player)
		if lobby.IsEmpty() {
			h.logger.Info("Lobby is empty, deleting lobby", "lobbyCode", lobby.Code)
			delete(h.lobbies, lobby.Code)
		}
	}
}
