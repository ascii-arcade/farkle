package wsclient

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ascii-arcade/farkle/internal/server"
	"golang.org/x/net/websocket"
)

type Client struct {
	url    string
	conn   *websocket.Conn
	logger *slog.Logger
}

func NewWsClient(logger *slog.Logger, url string) *Client {
	l := logger.With("component", "wsclient", "url", url)
	return &Client{
		url:    url,
		logger: l,
	}
}

func (c *Client) Connect() error {
	attempts := 0
TRYAGAIN:
	conn, err := websocket.Dial(c.url, "", "http://localhost/")
	if err != nil {
		if attempts < 5 {
			attempts++
			goto TRYAGAIN
		}
		return err
	}
	c.conn = conn

	go func() {
		for {
			if err := c.MonitorMessages(); err != nil {
				c.logger.Error("Error monitoring messages", "error", err)

				time.Sleep(5 * time.Second)
				c.logger.Debug("Attempting to reconnect...")

				if err := c.Connect(); err != nil {
					c.logger.Error("Error reconnecting", "error", err)
				}
				return
			}
		}
	}()

	return nil
}

func (c *Client) Connected() bool {
	if c.conn == nil {
		return false
	}
	return c.conn.IsServerConn()
}

func (c *Client) Close() {
	if err := c.conn.Close(); err != nil {
		fmt.Println("Error closing connection:", err)
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
			fmt.Println("Received ping from server")
			if err := c.Ping(); err != nil {
				return err
			}
		}
	}
}
