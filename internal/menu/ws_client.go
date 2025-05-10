package menu

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ascii-arcade/farkle/internal/message"
	"golang.org/x/net/websocket"
)

type client struct {
	conn      *websocket.Conn
	connected bool
	logger    *slog.Logger
	name      string

	disconnect chan bool
}

func newWsClient(name string) *client {
	l := logger.With("component", "wsclient")
	c := &client{
		logger:     l,
		name:       name,
		disconnect: make(chan bool),
	}
	return c.connect()
}

func (c *client) connect() *client {
	url := fmt.Sprintf("ws://%s:%s/ws?name=%s", serverURL, serverPort, c.name)
	c.logger.Debug("Connecting to server", "url", url)
	go func() {
		for {
			select {
			case <-c.disconnect:
				c.logger.Debug("Disconnecting from server")
				if err := c.conn.Close(); err != nil {
					c.logger.Error("Error closing connection", "error", err)
				}
				c.connected = false
				return
			default:
			}

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

func (c *client) SendMessage(msg message.Message) error {
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
	msg := message.Message{
		Channel: message.ChannelPing,
		Type:    message.MessageTypePing,
		SentAt:  time.Now(),
	}
	return c.SendMessage(msg)
}

func (c *client) monitorMessages() error {
	for {
		select {
		case <-c.disconnect:
			c.logger.Debug("shutting down message monitor")
			return nil
		default:
		}

		if c.conn == nil || !c.connected {
			c.logger.Error("Connection is nil, cannot monitor messages")
			return nil
		}

		msgRaw := make([]byte, 4096) // Allocate a buffer with a fixed size
		n, err := c.conn.Read(msgRaw)
		if err != nil {
			return err
		}

		var msg message.Message
		if err := json.Unmarshal(msgRaw[:n], &msg); err != nil {
			return err
		}

		switch msg.Channel {
		case message.ChannelPing:
			c.logger.Debug("Received ping from server")
			if err := c.ping(); err != nil {
				return err
			}
		case message.ChannelPlayer:
			c.logger.Debug("Received player message from server")
			switch msg.Type {
			case message.MessageTypeMe:
				c.logger.Debug("Received player message from server")
				if err = json.Unmarshal([]byte(msg.Data.(string)), &me); err != nil {
					c.logger.Error("Error unmarshalling player message", "error", err)
					continue
				}
			}
		case message.ChannelLobby:
			switch msg.Type {
			case message.MessageTypeCreated, message.MessageTypeUpdated:
				c.logger.Debug("Received lobby update from server")
				if err = json.Unmarshal([]byte(msg.Data.(string)), &currentLobby); err != nil {
					c.logger.Error("Error unmarshalling player message", "error", err)
					continue
				}
			}
		}
	}
}
