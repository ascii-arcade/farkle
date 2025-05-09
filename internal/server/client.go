package server

import (
	"time"

	"github.com/ascii-arcade/farkle/internal/player"
	"github.com/rs/xid"
	"golang.org/x/net/websocket"
)

type client struct {
	id       string
	active   bool
	lastSeen time.Time
	conn     *websocket.Conn
	player   *player.Player
}

func (h *hub) newClient(conn *websocket.Conn) *client {
	p := &player.Player{}

	c := &client{
		id:       xid.New().String(),
		conn:     conn,
		lastSeen: time.Now().Add(15 * time.Second),
		active:   true,
		player:   p,
	}
	h.register <- c
	go c.handleMessages(h)
	return c
}

func (c *client) handleMessages(h *hub) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg Message
		if err := websocket.JSON.Receive(c.conn, &msg); err != nil {
			h.logger.Error("Failed to receive message", "error", err)
			break
		}

		switch msg.Channel {
		case ChannelPing:
			h.logger.Info("Received ping from client", "clientId", c.id)
			c.lastSeen = time.Now()
		case ChannelLobby:
			h.logger.Info("Received lobby message from client", "clientId", c.id)
			switch msg.Type {
			case MessageTypeCreate:
				data := msg.Data.(map[string]any)
				host := player.New(data["hosts_name"].(string))
				c.player = host

				returnMsg := Message{
					Channel: ChannelPlayer,
					Type:    MessageTypeMe,
					Data:    host.ToBytes(),
					SentAt:  time.Now(),
				}

				if _, err := c.conn.Write(returnMsg.toBytes()); err != nil {
					h.logger.Error("Failed to send new player message", "error", err)
					continue
				}

				newLobby := h.createLobby(
					host,
					data["hosts_name"].(string),
				)
				returnMsg = Message{
					Channel: ChannelLobby,
					Type:    MessageTypeCreated,
					Data:    newLobby.ToBytes(),
					SentAt:  time.Now(),
				}
				c.conn.Write(returnMsg.toBytes())
			case MessageTypeJoin:
				// data := map[string]any{}
				// if err := json.Unmarshal(msg.Data, &data); err != nil {
				// 	h.logger.Error("Failed to unmarshal join message", "error", err)
				// 	continue
				// }
				// lobbyId, ok := data["lobby"].(string)
				// if !ok {
				// 	h.logger.Error("Invalid lobby ID in join message")
				// 	continue
				// }
				// lobby := h.getLobby(lobbyId)
				// if lobby == nil {
				// 	h.logger.Error("Lobby not found", "lobbyId", lobbyId)
				// 	continue
				// }
				// name, ok := data["name"].(string)
				// if !ok {
				// 	h.logger.Error("Invalid name in join message")
				// 	continue
				// }
				// newPlayer := lobby.AddPlayer(name)
				// if newPlayer == nil {
				// 	h.logger.Error("Failed to add player to lobby", "lobbyId", lobbyId, "name", name)
				// 	continue
				// }
				// c.player = newPlayer
				// newPlayerMsg := Message{
				// 	Channel: ChannelPlayer,
				// 	Type:    MessageTypeMe,
				// 	Data:    newPlayer.ToBytes(),
				// }
				// if _, err := c.conn.Write(newPlayerMsg.toBytes()); err != nil {
				// 	h.logger.Error("Failed to send new player message", "error", err)
				// 	continue
				// }
			}

		}
	}
}
