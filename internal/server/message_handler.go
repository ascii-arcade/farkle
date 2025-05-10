package server

import (
	"strings"
	"time"

	"github.com/ascii-arcade/farkle/internal/message"
	"github.com/ascii-arcade/farkle/internal/player"
)

func (h *hub) handleMessages(p *player.Player) {
	for {
		if p == nil {
			h.logger.Error("Player is nil, no longer watching for messages")
			return
		}

		msg, err := p.ReceiveMessage()
		if err != nil {
			if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "use of closed network connection") {
				h.logger.Info("Client disconnected", "clientId", p.Id)
				h.removePlayer(p)
				return
			}
			h.logger.Error("Failed to receive message", "error", err)
			continue
		}

		switch msg.Channel {
		case message.ChannelPing:
			h.logger.Debug("Received ping from client", "clientId", p.Id)
			p.LastSeen = time.Now()
		case message.ChannelLobby:
			h.logger.Debug("Received lobby message from client", "clientId", p.Id)
			switch msg.Type {
			case message.MessageTypeCreate:
				returnMsg := message.Message{
					Channel: message.ChannelPlayer,
					Type:    message.MessageTypeMe,
					Data:    p.ToJSON(),
					SentAt:  time.Now(),
				}

				if err := p.SendMessage(returnMsg); err != nil {
					h.logger.Error("Failed to send new player message", "error", err)
					continue
				}

				newLobby := h.createLobby(p)
				returnMsg = message.Message{
					Channel: message.ChannelLobby,
					Type:    message.MessageTypeCreated,
					Data:    newLobby.ToJSON(),
					SentAt:  time.Now(),
				}
				if err := p.SendMessage(returnMsg); err != nil {
					h.logger.Error("Failed to send new lobby message", "error", err)
					continue
				}
			case message.MessageTypeJoin:
				code := msg.Data.(string)
				lobby := h.getLobby(code)
				if lobby == nil {
					h.logger.Error("Lobby not found", "lobbyCode", code)
					continue
				}

				if ok := lobby.AddPlayer(p); !ok {
					h.logger.Error("Failed to add player to lobby", "error", err)
					continue
				}

				returnMsg := message.Message{
					Channel: message.ChannelLobby,
					Type:    message.MessageTypeUpdated,
					Data:    lobby.ToJSON(),
					SentAt:  time.Now(),
				}
				h.broadcastMessage(returnMsg, lobby.Players...)

			case message.MessageTypeLeave:
				h.logger.Info("Client left lobby", "clientId", p.Id)
				h.removePlayerFromLobby(p)
			}

		}
	}
}
