package lobbies

import (
	"time"

	"github.com/ascii-arcade/farkle/internal/message"
)

func (l *Lobby) BroadcastUpdate() {
	for _, player := range l.Players {
		if player != nil {
			player.SendMessage(message.Message{
				Channel: message.ChannelLobby,
				Type:    message.MessageTypeUpdated,
				Data:    l.ToJSON(),
				SentAt:  time.Now(),
			})
		}
	}
}

func (l *Lobby) broadcastGameUpdate(rolled bool) {
	for _, player := range l.Players {
		if player != nil {
			msg := message.Message{
				Channel: message.ChannelGame,
				Type:    message.MessageTypeUpdated,
				Data:    l.Game.ToJSON(),
				SentAt:  time.Now(),
			}

			if rolled {
				msg.Type = message.MessageTypeRolled
			}
			player.SendMessage(msg)
		}
	}
}
