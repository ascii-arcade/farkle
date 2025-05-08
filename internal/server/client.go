package server

import (
	"encoding/json"
	"time"

	"github.com/ascii-arcade/farkle/internal/lobby"
	"github.com/rs/xid"
	"golang.org/x/net/websocket"
)

type client struct {
	id       string
	active   bool
	lastSeen time.Time
	conn     *websocket.Conn
}

func (h *hub) newClient(conn *websocket.Conn) *client {
	c := &client{
		id:       xid.New().String(),
		conn:     conn,
		lastSeen: time.Now().Add(15 * time.Second),
		active:   true,
	}
	h.register <- c
	go c.handleMessages(h)
	return c
}

func (c *client) handleMessages(h *hub) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg Message
		if err := websocket.JSON.Receive(c.conn, &msg); err != nil {
			h.logger.Error("Failed to receive message", "error", err)
			break
		}

		switch msg.Channel {
		case ChannelPing:
			h.logger.Info("Received ping from client", "clientId", c.id)
			c.lastSeen = time.Now()
		case ChannelLobby:
			h.logger.Info("Received lobby message from client", "clientId", c.id)
			if msg.Type == MessageTypeCreate {
				lobby := &lobby.Lobby{}
				if err := json.Unmarshal(msg.Data, lobby); err != nil {
					h.logger.Error("Failed to unmarshal lobby", "error", err)
					continue
				}
				h.addLobby(lobby)
			}
		}
	}
}
