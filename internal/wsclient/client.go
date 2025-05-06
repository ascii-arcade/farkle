package wsclient

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ascii-arcade/farkle/internal/server"
	"golang.org/x/net/websocket"
)

type WsClient struct {
	conn *websocket.Conn
}

func NewWsClient(url string) (*WsClient, error) {
	conn, err := websocket.Dial(url, "", "http://localhost/")
	if err != nil {
		return nil, err
	}

	return &WsClient{
		conn: conn,
	}, nil
}

func (c *WsClient) Close() {
	if err := c.conn.Close(); err != nil {
		fmt.Println("Error closing connection:", err)
	}
}

func (c *WsClient) Reconnect(url string) error {
	conn, err := websocket.Dial(url, "", "http://localhost/")
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *WsClient) SendMessage(msg server.Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := c.conn.Write(b); err != nil {
		return err
	}

	return nil
}

func (c *WsClient) Ping() error {
	msg := server.Message{
		Channel: server.ChannelPing,
		SentAt:  time.Now(),
	}
	return c.SendMessage(msg)
}

func (c *WsClient) MonitorMessages() error {
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
