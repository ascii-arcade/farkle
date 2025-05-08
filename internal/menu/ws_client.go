package menu

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/ascii-arcade/farkle/internal/lobby"
	"github.com/ascii-arcade/farkle/internal/player"
	"github.com/ascii-arcade/farkle/internal/server"
	"golang.org/x/net/websocket"
)

type client struct {
	url       string
	conn      *websocket.Conn
	connected bool
	logger    *slog.Logger
}

func newWsClient(logger *slog.Logger, url string) *client {
	l := logger.With("component", "wsclient", "url", url)
	c := &client{
		url:    url,
		logger: l,
	}
	c.connect()
	return c
}

func (c *client) connect() {
	c.logger.Debug("Connecting to server", "url", c.url)
	go func() {
		for {
			conn, err := websocket.Dial(c.url, "", "http://localhost/")
			if err != nil {
				c.logger.Error("Error connecting to server", "error", err)
				goto RECONNECT
			}
			c.connected = true
			c.conn = conn

			if err := c.monitorMessages(); err != nil {
				c.logger.Error("Error monitoring messages", "error", err)
			}

		RECONNECT:
			c.connected = false
			c.logger.Debug("Retrying connection...")
			time.Sleep(5 * time.Second)
		}
	}()
}

func (c *client) IsConnected() bool {
	return c.connected
}

func (c *client) Close() {
	if err := c.conn.Close(); err != nil {
		c.logger.Error("Error closing connection", "error", err)
	}
}

func (c *client) SendMessage(msg server.Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := c.conn.Write(b); err != nil {
		return err
	}

	return nil
}

func (c *client) ping() error {
	msg := server.Message{
		Channel: server.ChannelPing,
		SentAt:  time.Now(),
	}
	return c.SendMessage(msg)
}

func (c *client) monitorMessages() error {
	for {
		msgRaw := make([]byte, 1024) // Allocate a buffer with a fixed size
		n, err := c.conn.Read(msgRaw)
		if err != nil {
			return err
		}

		var msg server.Message
		if err := json.Unmarshal(msgRaw[:n], &msg); err != nil {
			return err
		}

		switch msg.Channel {
		case server.ChannelPing:
			c.logger.Debug("Received ping from server")
			if err := c.ping(); err != nil {
				return err
			}
		case server.ChannelLobby:
			switch msg.Type {
			case server.MessageTypeList:
				c.logger.Debug("Received lobby list from server")
				var l []*lobby.Lobby
				if err := json.Unmarshal(msg.Data, &l); err != nil {
					c.logger.Error("Error unmarshalling lobby list", "error", err)
					continue
				}
				for _, lobby := range l {
					updateLobby(lobby)
				}
			}
		case server.ChannelPlayer:
			switch msg.Type {
			case server.MessageTypeMe:
				c.logger.Debug("Received player info from server")
				var player *player.Player
				if err := json.Unmarshal(msg.Data, &player); err != nil {
					c.logger.Error("Error unmarshalling player info", "error", err)
					continue
				}
				myPlayer = player
			}
		}
	}
}
