package server

import (
	"time"
)

func (h *hub) monitorConnections() {
	for {
		time.Sleep(5 * time.Second)

		for p := range h.players {
			if time.Since(p.LastSeen) > 10*time.Second {
				h.mu.Lock()
				h.logger.Info("client inactive, closing connection", "player_id", p.Id)
				p.Close()
				delete(h.players, p)
				h.mu.Unlock()
			}
		}
	}
}
