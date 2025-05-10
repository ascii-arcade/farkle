package server

import (
	"time"

	"github.com/ascii-arcade/farkle/internal/message"
)

func (h *hub) monitorLobbies() {
	for {
		h.mu.Lock()
		for _, lobby := range h.lobbies {
			if lobby.IsEmpty() {
				h.logger.Info("Lobby is empty, removing", "lobbyId", lobby.Id)
				delete(h.lobbies, lobby.Id)
			}

			h.broadcastMessage(message.Message{
				Channel: message.ChannelLobby,
				Type:    message.MessageTypeUpdated,
				Data:    lobby.ToJSON(),
				SentAt:  time.Now(),
			}, lobby.Players...)
		}
		h.mu.Unlock()

		time.Sleep(1 * time.Second)
	}
}
