package lobbies

import "github.com/ascii-arcade/farkle/internal/message"

func (l *Lobby) handleMessages() {
	for msg := range l.messages {
		player := l.getPlayer(msg.PlayerId)

		l.logger.Debug("Received message from player", "channel", msg.Channel, "type", msg.Type, "player_id", msg.PlayerId, "player_name", player.Name)

		switch msg.Channel {
		case message.ChannelLobby:
			switch msg.Type {
			case message.MessageTypeStart:
				if player.Host {
					l.StartGame()
				}
			}
		case message.ChannelGame:
		}
	}
}

// func (h *hub) handleMessages(p *player.Player) {
// 	for {
// 		if p == nil {
// 			h.logger.Error("Player is nil, no longer watching for messages")
// 			return
// 		}

// 		msg, err := p.ReceiveMessage()
// 		if err != nil {
// 			if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "use of closed network connection") {
// 				h.logger.Info("Player disconnected", "player_id", p.Id)
// 				h.removePlayer(p)
// 				return
// 			}
// 			h.logger.Error("Failed to receive message", "error", err)
// 			continue
// 		}

// 		h.logger.Debug("Received message from player", "player_id", p.Id, "channel", msg.Channel, "type", msg.Type)

// 		switch msg.Channel {
// 		case message.ChannelPing:
// 			h.logger.Debug("Received ping from player", "player_id", p.Id)
// 			p.LastSeen = time.Now()
// 		case message.ChannelLobby:
// 			switch msg.Type {
// 			case message.MessageTypeCreate:
// 				newLobby := h.createLobby(p)
// 				returnMsg := message.Message{
// 					Channel: message.ChannelLobby,
// 					Type:    message.MessageTypeCreated,
// 					Data:    newLobby.ToJSON(),
// 					SentAt:  time.Now(),
// 				}
// 				if err := h.broadcastMessage(returnMsg, newLobby.Players...); err != nil {
// 					h.logger.Error("Failed to broadcast lobby message", "error", err)
// 					continue
// 				}
// 			case message.MessageTypeJoin:
// 				code := msg.Data.(string)
// 				lobby := h.getLobby(code)
// 				if lobby == nil {
// 					h.logger.Error("Lobby not found", "lobbyCode", code)
// 					continue
// 				}

// 				if lobby.Started {
// 					returnMsg := message.Message{
// 						Channel: message.ChannelLobby,
// 						Type:    message.MessageTypeError,
// 						Data:    "Lobby already started",
// 						SentAt:  time.Now(),
// 					}
// 					if err := p.SendMessage(returnMsg); err != nil {
// 						h.logger.Error("Failed to send error message", "error", err)
// 					}
// 					continue
// 				}

// 				if ok := lobby.AddPlayer(p); !ok {
// 					h.logger.Error("Failed to add player to lobby", "error", err)
// 					continue
// 				}

// 				returnMsg := message.Message{
// 					Channel: message.ChannelLobby,
// 					Type:    message.MessageTypeUpdated,
// 					Data:    lobby.ToJSON(),
// 					SentAt:  time.Now(),
// 				}
// 				h.broadcastMessage(returnMsg, lobby.Players...)

// 			case message.MessageTypeLeave:
// 				h.logger.Info("Client left lobby", "clientId", p.Id)
// 				h.removePlayerFromLobby(p)
// 			case message.MessageTypeStart:
// 				lobby := h.getLobby(msg.Data.(string))
// 				if lobby == nil {
// 					h.logger.Error("Lobby not found", "lobbyCode", msg.Data.(string))
// 					continue
// 				}
// 				if lobby.Started {
// 					returnMsg := message.Message{
// 						Channel: message.ChannelLobby,
// 						Type:    message.MessageTypeError,
// 						Data:    "Lobby already started",
// 						SentAt:  time.Now(),
// 					}
// 					if err := p.SendMessage(returnMsg); err != nil {
// 						h.logger.Error("Failed to send error message", "error", err)
// 					}
// 					continue
// 				}

// 				lobby.StartGame()
// 				returnMsg := message.Message{
// 					Channel: message.ChannelLobby,
// 					Type:    message.MessageTypeStarted,
// 					Data:    lobby.ToJSON(),
// 					SentAt:  time.Now(),
// 				}
// 				h.broadcastMessage(returnMsg, lobby.Players...)
// 			}
// 		case message.ChannelGame:
// 			gameDetails := game.GameDetails{}
// 			if err := json.Unmarshal([]byte(msg.Data.(string)), &gameDetails); err != nil {
// 				h.logger.Error("Failed to unmarshal game message", "error", err)
// 				continue
// 			}

// 			lobby := h.getLobby(gameDetails.LobbyCode)
// 			if lobby == nil {
// 				h.logger.Error("Lobby not found", "lobbyCode", gameDetails.LobbyCode)
// 				continue
// 			}
// 			if lobby.Game == nil {
// 				h.logger.Error("Game not found", "lobbyCode", gameDetails.LobbyCode)
// 				continue
// 			}

// 			lobby.Game.HandleMessage(msg)
// 			returnMsg := message.Message{
// 				Channel: message.ChannelGame,
// 				Type:    message.MessageTypeUpdated,
// 				Data:    lobby.Game.ToJSON(),
// 				SentAt:  time.Now(),
// 			}

// 			if msg.Type == message.MessageTypeRoll {
// 				returnMsg.Type = message.MessageTypeRolled
// 			}
// 			h.broadcastMessage(returnMsg, lobby.Players...)
// 		}
// 	}
// }
