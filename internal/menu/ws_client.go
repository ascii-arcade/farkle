package menu

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ascii-arcade/farkle/internal/server"
	"golang.org/x/net/websocket"
)

type client struct {
	conn      *websocket.Conn
	connected bool
	logger    *slog.Logger
}

func newWsClient(logger *slog.Logger) *client {
	l := logger.With("component", "wsclient")
	c := &client{
		logger: l,
	}
	return c.connect()
}

func (c *client) connect() *client {
	url := fmt.Sprintf("ws://%s:%s/ws", serverURL, serverPort)
	c.logger.Debug("Connecting to server", "url", url)
	go func() {
		for {
			conn, err := websocket.Dial(url, "", "http://localhost/")
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
	return c
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
	if !c.connected {
		c.waitForConnection()
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := c.conn.Write(b); err != nil {
		return err
	}

	return nil
}

func (c *client) waitForConnection() {
	for !c.connected {
		time.Sleep(1 * time.Second)
	}
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
		case server.ChannelPlayer:
			c.logger.Debug("Received player message from server")
			switch msg.Type {
			case server.MessageTypeMe:
				c.logger.Debug("Received player message from server")
				if err := json.Unmarshal(msg.Data.([]byte), &me); err != nil {
					c.logger.Error("Error unmarshalling player message", "error", err)
					continue
				}
			}
		case server.ChannelLobby:
			switch msg.Type {
			case server.MessageTypeUpdated:
				c.logger.Debug("Received lobby list from server")
				if err := json.Unmarshal(msg.Data.([]byte), &currentLobby); err != nil {
					c.logger.Error("Error unmarshalling lobby list", "error", err)
					continue
				}
			case server.MessageTypeCreated:
				c.logger.Debug("Received lobby created message from server")
				if err := json.Unmarshal(msg.Data.([]byte), &currentLobby); err != nil {
					c.logger.Error("Error unmarshalling lobby created message", "error", err)
					continue
				}
			}
		}
	}
}
