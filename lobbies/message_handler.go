package lobbies

import (
	"encoding/json"

	"github.com/ascii-arcade/farkle/game"
	"github.com/ascii-arcade/farkle/message"
)

func (l *Lobby) handleMessages() {
	for msg := range l.messages {
		player := l.getPlayer(msg.PlayerId)

		l.logger.Info("Received message from player", "channel", msg.Channel, "type", msg.Type, "player_id", msg.PlayerId, "player_name", player.Name)

		switch msg.Channel {
		case message.ChannelLobby:
			switch msg.Type {
			case message.MessageTypeStart:
				if player.Host {
					l.StartGame()
					l.BroadcastUpdate()
				}
			}
		case message.ChannelGame:
			if l.Game == nil {
				l.logger.Error("Game not found", "lobbyCode", l.Code)
				continue
			}
			details := game.GameDetails{}
			if err := json.Unmarshal([]byte(msg.Data), &details); err != nil {
				return
			}
			rolled := false
			switch msg.Type {
			case message.MessageTypeRoll:
				l.Game.RollDice()
				rolled = true
			case message.MessageTypeHold:
				l.Game.HoldDie(details.DieHeld)
			case message.MessageTypeUndo:
				l.Game.Undo()
			case message.MessageTypeLock:
				l.Game.LockDice()
			}

			l.broadcastGameUpdate(rolled)
		}
	}
}
