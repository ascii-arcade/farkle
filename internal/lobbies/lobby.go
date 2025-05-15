package lobbies

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/ascii-arcade/farkle/internal/game"
	"github.com/ascii-arcade/farkle/internal/message"
	"github.com/ascii-arcade/farkle/internal/player"
)

type Lobby struct {
	Code      string           `json:"code"`
	Players   []*player.Player `json:"players"`
	Started   bool             `json:"started"`
	CreatedAt time.Time        `json:"created_at"`
	Game      *game.Game       `json:"game"`

	messages chan message.Message `json:"-"`

	logger *slog.Logger
}

func NewLobby(logger *slog.Logger) *Lobby {
	code := newCode()
	l := &Lobby{
		Code:      code,
		Players:   make([]*player.Player, 6),
		Started:   false,
		CreatedAt: time.Now(),
		logger:    logger.With("lobby_code", code),
		messages:  make(chan message.Message, 100),
	}

	go l.handleMessages()

	return l
}

func (l *Lobby) AddPlayer(p *player.Player) bool {
	for i, player := range l.Players {
		if player != nil {
			if player.Id == p.Id {
				l.logger.Debug("Player already in lobby", "player_id", p.Id)
				return true
			}

			continue
		}
		if i == 0 {
			p.Host = true
		}

		go p.MonitorMessages(l.messages)

		l.Players[i] = p

		return true
	}

	return false
}

func (l *Lobby) RemovePlayer(p *player.Player) {
	for i, player := range l.Players {
		if player == nil {
			continue
		}

		if player.Id == p.Id {
			if player.Host {
				l.NewHost()
			}

			l.Players[i] = nil
			return
		}
	}
}

func (l *Lobby) NewHost() {
	for i, player := range l.Players {
		if player != nil && !player.Host {
			l.Players[i].Host = true
			return
		}
	}
}

func (l *Lobby) Ready() bool {
	playerCount := 0
	for _, player := range l.Players {
		if player != nil {
			playerCount++
			if playerCount >= 2 {
				return true
			}
		}
	}

	return false
}

func (l *Lobby) ToJSON() string {
	b, err := json.Marshal(l)
	if err != nil {
		return ""
	}
	return string(b)
}

func (l *Lobby) GetHost() *player.Player {
	for _, player := range l.Players {
		if player != nil && player.Host {
			return player
		}
	}
	return nil
}

func (l *Lobby) IsHost(p *player.Player) bool {
	return l.GetHost().Id == p.Id
}

func (l *Lobby) IsEmpty() bool {
	for _, player := range l.Players {
		if player != nil {
			return false
		}
	}
	return true
}

func (l *Lobby) IsFull() bool {
	for _, player := range l.Players {
		if player == nil {
			return false
		}
	}
	return true
}

func (l *Lobby) StartGame() {
	l.Game = game.New(l.Code, l.Players)
	l.Started = true
}

func (l *Lobby) getPlayer(id string) *player.Player {
	for _, p := range l.Players {
		if p != nil && p.Id == id {
			return p
		}
	}
	return nil
}
