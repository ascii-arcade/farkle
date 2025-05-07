package wsclient

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/ascii-arcade/farkle/internal/server"
	"golang.org/x/net/websocket"
)

type Client struct {
	url       string
	conn      *websocket.Conn
	connected bool
	logger    *slog.Logger
}

func NewWsClient(logger *slog.Logger, url string) *Client {
	l := logger.With("component", "wsclient", "url", url)
	return &Client{
		url:    url,
		logger: l,
	}
}

func (c *Client) Connect() {
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

			if err := c.MonitorMessages(); err != nil {
				c.logger.Error("Error monitoring messages", "error", err)
			}

		RECONNECT:
			c.connected = false
			c.logger.Debug("Retrying connection...")
			time.Sleep(5 * time.Second)
		}
	}()
}

func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) Close() {
	if err := c.conn.Close(); err != nil {
		c.logger.Error("Error closing connection", "error", err)
	}
}

func (c *Client) Reconnect(url string) error {
	conn, err := websocket.Dial(url, "", "http://localhost/")
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) SendMessage(msg server.Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := c.conn.Write(b); err != nil {
		return err
	}

	return nil
}

func (c *Client) Ping() error {
	msg := server.Message{
		Channel: server.ChannelPing,
		SentAt:  time.Now(),
	}
	return c.SendMessage(msg)
}

func (c *Client) MonitorMessages() error {
	for {
		var msgRaw = make([]byte, 512)
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
			if err := c.Ping(); err != nil {
				return err
			}
		}
	}
}
