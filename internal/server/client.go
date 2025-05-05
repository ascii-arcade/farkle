package server

import (
	"time"

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
		var msg message
		if err := websocket.JSON.Receive(c.conn, &msg); err != nil {
			h.logger.Error("Failed to receive message", "error", err)
			break
		}

		switch msg.channel {
		case ChannelPing:
			h.logger.Info("Received ping from client", "clientId", c.id)
			c.lastSeen = time.Now()
		}
	}
}
