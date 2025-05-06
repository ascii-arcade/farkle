package server

import "time"

func (h *hub) monitorConnections() {
	for {
		if len(h.clients) > 0 {
			h.broadcast <- Message{
				Channel: ChannelPing,
				SentAt:  time.Now(),
			}
		}

		time.Sleep(5 * time.Second)

		for c := range h.clients {
			if time.Since(c.lastSeen) > 10*time.Second {
				h.mu.Lock()
				h.logger.Info("client inactive, closing connection", "client", c)
				c.conn.Close()
				delete(h.clients, c)
				h.mu.Unlock()
			}
		}
	}
}
