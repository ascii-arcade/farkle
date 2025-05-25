package networkmanager

import (
	"fmt"

	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/message"
	"golang.org/x/net/websocket"
)

type NetworkManager struct {
	Incoming chan message.Message
	Outgoing chan message.Message

	conn *websocket.Conn
	quit chan struct{}
}

func NewNetworkManager(code, name string) (*NetworkManager, error) {
	scheme := "ws"
	if config.GetSecure() {
		scheme = "wss"
	}

	port := ""
	if config.GetServerPort() != "" {
		port = ":" + config.GetServerPort()
	}

	url := fmt.Sprintf("%s://%s%s/ws/%s?name=%s", scheme, config.GetServerURL(), port, code, name)
	wsConfig, err := websocket.NewConfig(url, "http://localhost/")
	if err != nil {
		return nil, err
	}
	wsConfig.Header.Set("Connection", "Upgrade")
	wsConfig.Header.Set("Upgrade", "websocket")
	conn, err := websocket.DialConfig(wsConfig)
	if err != nil {
		return nil, err
	}

	nm := &NetworkManager{
		Incoming: make(chan message.Message),
		Outgoing: make(chan message.Message),
		conn:     conn,
		quit:     make(chan struct{}),
	}

	go nm.readMessages()
	go nm.writeMessages()

	return nm, nil
}

func (nm *NetworkManager) readMessages() {
	for {
		select {
		case <-nm.quit:
			return
		default:
			var msg message.Message
			if err := websocket.JSON.Receive(nm.conn, &msg); err != nil {
				close(nm.Incoming)
				return
			}
			nm.Incoming <- msg
		}
	}
}

func (nm *NetworkManager) writeMessages() {
	for {
		select {
		case <-nm.quit:
			return
		case msg := <-nm.Outgoing:
			if err := websocket.JSON.Send(nm.conn, msg); err != nil {
				close(nm.Outgoing)
				return
			}
		}
	}
}

func (nm *NetworkManager) Close() error {
	close(nm.quit)
	if err := nm.conn.Close(); err != nil {
		return err
	}
	return nil
}
