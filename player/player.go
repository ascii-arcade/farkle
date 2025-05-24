package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/message"
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

	lastSeen   time.Time       `json:"-"`
	conn       *websocket.Conn `json:"-"`
	logger     *slog.Logger    `json:"-"`
	disconnect chan bool       `json:"-"`
	// player   *player.Player `json:"-"`
}

func NewPlayer(logger *slog.Logger, conn *websocket.Conn, name string) *Player {
	c := &Player{
		Id:       xid.New().String(),
		Name:     name,
		Active:   true,
		lastSeen: time.Now().Add(15 * time.Second),

		conn:   conn,
		logger: logger,
	}
	return c
}

func (p *Player) Connect(code string, messageChan chan message.Message) {
	url := fmt.Sprintf("ws://%s:%s/ws/%s?name=%s", config.GetServerURL(), config.GetServerPort(), code, p.Name)
	p.logger.Debug("connecting to server", "url", url)
	go func() {
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

			if err := p.MonitorMessages(messageChan); err != nil && !p.Active {
				p.logger.Error("error monitoring messages", "error", err)
			}

		RECONNECT:
			p.Active = false
			p.logger.Debug("retrying connection...")
			time.Sleep(1 * time.Second)
		}
	}()
}

func (p *Player) Disconnected() chan bool {
	return p.disconnect
}

func (p *Player) Update(pIn Player) {
	p.Name = pIn.Name
	p.Score = pIn.Score
	p.Host = pIn.Host
	p.Active = pIn.Active
	p.Color = pIn.Color
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

func (p *Player) Close() {
	if p == nil || p.conn == nil {
		return
	}
	_ = p.conn.Close()
	p.Active = false
}

func (p *Player) Connected() bool {
	return p.conn != nil
}

func (p *Player) StyledPlayerName() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Color))

	return style.Render(p.Name)
}
