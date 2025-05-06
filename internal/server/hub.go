package server

import (
	"log/slog"
	"sync"
)

type hub struct {
	clients    map[*client]bool
	broadcast  chan Message
	register   chan *client
	unregister chan *client

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
		if _, err := c.conn.Write(msg.toBytes()); err != nil {
			logger.Error("Failed to send message", "error", err)
			c.conn.Close()
			delete(h.clients, c)
		}
	}
}
