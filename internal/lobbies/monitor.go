package lobbies

import (
	"log/slog"
	"time"
)

func MonitorLobbies(l *slog.Logger) {
	logger := l.With("monitor", "lobbies")

	for {
		mu.Lock()
		for _, lobby := range lobbies {
			if lobby.IsEmpty() {
				logger.Info("Lobby is empty, removing", "lobby_code", lobby.Code)
				delete(lobbies, lobby.Code)
				continue
			}

			// if lobby.Started {
			// 	if err := h.broadcastMessage(message.Message{
			// 		Channel: message.ChannelGame,
			// 		Type:    message.MessageTypeUpdated,
			// 		Data:    lobby.Game.ToJSON(),
			// 		SentAt:  time.Now(),
			// 	}, lobby.Players...); err != nil {
			// 		h.logger.Error("Failed to broadcast game message", "error", err)
			// 	}
			// }

			// if err := h.broadcastMessage(message.Message{
			// 	Channel: message.ChannelLobby,
			// 	Type:    message.MessageTypeUpdated,
			// 	Data:    lobby.ToJSON(),
			// 	SentAt:  time.Now(),
			// }, lobby.Players...); err != nil {
			// 	h.logger.Error("Failed to broadcast lobby message", "error", err)
			// }
		}
		mu.Unlock()

		time.Sleep(1 * time.Second)
	}
}
