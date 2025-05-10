package server

import (
	"time"

	"github.com/ascii-arcade/farkle/internal/message"
)

func (h *hub) monitorConnections() {
	for {
		if len(h.players) > 0 {
			h.broadcastMessage(message.Message{
				Channel: message.ChannelPing,
				Type:    message.MessageTypePing,
				SentAt:  time.Now(),
			})
		}

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
