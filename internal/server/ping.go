package server

import "time"

func (h *hub) monitorConnections() {
	for {
		if len(h.clients) > 0 {
			h.broadcast <- message{
				channel: ChannelPing,
				sentAt:  time.Now(),
			}
		}

		time.Sleep(5 * time.Second)

		for c := range h.clients {
			if time.Since(c.lastSeen) > 10*time.Second {
				h.logger.Info("client inactive, closing connection", "client", c)
				c.conn.Close()
				delete(h.clients, c)
			}
		}
	}
}
