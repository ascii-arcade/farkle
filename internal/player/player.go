package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ascii-arcade/farkle/internal/config"
	"github.com/ascii-arcade/farkle/internal/message"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/xid"
	"golang.org/x/net/websocket"
)

type Player struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Score  int    `json:"score"`
	Host   bool   `json:"host"`
	Active bool   `json:"active"`
	Color  string `json:"color"`

	LastSeen       time.Time       `json:"-"`
	conn           *websocket.Conn `json:"-"`
	logger         *slog.Logger    `json:"-"`
	disconnect     chan bool
	lobbyMessages  chan message.Message
	gameMessages   chan message.Message
	playerMessages chan message.Message
}

func NewPlayer(conn *websocket.Conn, logger *slog.Logger, name string) *Player {
	if logger == nil {
		logger = slog.Default()
	}
	c := &Player{
		Id:       xid.New().String(),
		Name:     name,
		Active:   true,
		LastSeen: time.Now().Add(15 * time.Second),

		conn:   conn,
		logger: logger.With("component", "player"),
	}
	return c
}

func (p *Player) Connect() {
	url := fmt.Sprintf("ws://%s:%s/ws?name=%s", config.GetServerURL(), config.GetServerPort(), p.Name)
	p.logger.Debug("connecting to server", "url", url)
	for {
		select {
		case <-p.disconnect:
			p.logger.Debug("disconnected from server")
			return
		default:
		}

		conn, err := websocket.Dial(url, "", "http://localhost/")
		if err != nil {
			p.logger.Error("error connecting to server", "error", err)
			goto RECONNECT
		}
		p.Active = true
		p.conn = conn

		if err := p.monitorMessages(); err != nil && !p.Active {
			p.logger.Error("error monitoring messages", "error", err)
		}

	RECONNECT:
		p.Active = false
		p.logger.Debug("retrying connection...")
		time.Sleep(1 * time.Second)
	}
}

func (p *Player) Close() {
	close(p.disconnect)

	if err := p.conn.Close(); err != nil {
		p.logger.Error("error closing connection", "error", err)
		return
	}
	close(LobbyMessages)
	close(GameMessages)
	close(PlayerMessages)
	c.connected = false
	c.logger.Debug("closed connection to server")
}

func (p *Player) ToJSON() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func (p *Player) SendMessage(msg message.Message) error {
	if p.conn == nil {
		return nil
	}
	if err := websocket.JSON.Send(p.conn, msg); err != nil {
		return err
	}
	return nil
}

func (p *Player) ReceiveMessage() (message.Message, error) {
	var msg message.Message
	if p.conn == nil {
		return message.Message{}, errors.New("no connection")
	}
	if err := websocket.JSON.Receive(p.conn, &msg); err != nil {
		return message.Message{}, err
	}
	return msg, nil
}

func (p *Player) Close() error {
	if p.conn == nil {
		return nil
	}
	if err := p.conn.Close(); err != nil {
		return err
	}
	p.Active = false
	return nil
}

func (p *Player) Connected() bool {
	return p.conn != nil
}

func (p *Player) styledPlayerName(i int) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Color))

	return style.Render(p.Name)
}
