package wsclient

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ascii-arcade/farkle/internal/config"
	"github.com/ascii-arcade/farkle/internal/message"
	"golang.org/x/net/websocket"
)

type Client struct {
	*websocket.Conn
	connected bool
	logger    *slog.Logger
	name      string
	id        string
}

var (
	client *Client

	LobbyMessages  chan message.Message
	GameMessages   chan message.Message
	PlayerMessages chan message.Message
	Disconnect     chan bool
)

func New(logger *slog.Logger, name string) {
	l := logger.With("component", "wsclient")
	c := &Client{
		logger: l,
		name:   name,
	}
	Disconnect = make(chan bool)
	LobbyMessages = make(chan message.Message, 10)
	GameMessages = make(chan message.Message, 10)
	PlayerMessages = make(chan message.Message, 10)

	go c.keepAlive()
	go c.connect()

	client = c
}

func (c *Client) connect() {
	url := fmt.Sprintf("ws://%s:%s/ws?name=%s", config.GetServerURL(), config.GetServerPort(), c.name)
	c.logger.Debug("connecting to server", "url", url)
	for {
		select {
		case <-Disconnect:
			c.logger.Debug("disconnected from server")
			return
		default:
		}

		conn, err := websocket.Dial(url, "", "http://localhost/")
		if err != nil {
			c.logger.Error("error connecting to server", "error", err)
			goto RECONNECT
		}
		c.connected = true
		c.Conn = conn

		if err := c.monitorMessages(); err != nil && !c.connected {
			c.logger.Error("error monitoring messages", "error", err)
		}

	RECONNECT:
		c.connected = false
		c.logger.Debug("retrying connection...")
		time.Sleep(1 * time.Second)
	}
}

func GetClient() *Client {
	return client
}

func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) Close() {
	if c == nil {
		return
	}
	close(Disconnect)

	if err := c.Conn.Close(); err != nil {
		c.logger.Error("error closing connection", "error", err)
		return
	}
	close(LobbyMessages)
	close(GameMessages)
	close(PlayerMessages)
	c.connected = false
	c.logger.Debug("closed connection to server")
}

func SendMessage(msg message.Message) error {
	if !client.connected {
		client.waitForConnection()
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return SendRawMessage(b)
}

func SendRawMessage(msg []byte) error {
	if !client.connected {
		client.waitForConnection()
	}

	if _, err := client.Write(msg); err != nil {
		return err
	}

	return nil
}

func (c *Client) keepAlive() {
	for {
		select {
		case <-Disconnect:
			c.logger.Debug("shutting down keep alive")
			return
		default:
		}

		if err := ping(); err != nil {
			c.logger.Error("error sending ping", "error", err)
		}

		time.Sleep(5 * time.Second)
	}
}

func (c *Client) waitForConnection() {
	for !c.connected {
		time.Sleep(1 * time.Second)
	}
}

func ping() error {
	msg := message.Message{
		Channel: message.ChannelPing,
		SentAt:  time.Now(),
	}
	return SendMessage(msg)
}

func (c *Client) monitorMessages() error {
	for {
		select {
		case <-Disconnect:
			c.logger.Debug("shutting down message monitor")
			return nil
		default:
		}

		if c.Conn == nil || !c.connected {
			c.logger.Error("connection is nil, cannot monitor messages")
			return nil
		}

		msgRaw := make([]byte, 4096) // Allocate a buffer with a fixed size
		n, err := c.Read(msgRaw)
		if err != nil {
			return err
		}

		var msg message.Message
		if err := json.Unmarshal(msgRaw[:n], &msg); err != nil {
			return err
		}

		c.logger.Debug("received message from server", "channel", msg.Channel, "type", msg.Type)

		switch msg.Channel {
		case message.ChannelPlayer:
			c.logger.Debug("received player message from server")
			PlayerMessages <- msg
		case message.ChannelLobby:
			c.logger.Debug("received lobby message from server")
			LobbyMessages <- msg
		case message.ChannelGame:
			c.logger.Debug("received game message from server")
			GameMessages <- msg
		}
	}
}
