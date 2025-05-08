package lobby

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/ascii-arcade/farkle/internal/tui"
	"github.com/rs/xid"
)

type Lobby struct {
	Id        string
	Name      string
	Players   []*tui.Player
	Started   bool
	CreatedAt time.Time
}

func NewLobby(lobbyName, hostName string, playerCount int) *Lobby {
	players := make([]*tui.Player, playerCount)
	players[0] = &tui.Player{
		Name:  hostName,
		Score: 0,
		Host:  true,
	}

	return &Lobby{
		Id:        xid.New().String(),
		Name:      lobbyName,
		Players:   players,
		Started:   false,
		CreatedAt: time.Now(),
	}
}

func (l *Lobby) AddPlayer(name string) {}

func (l *Lobby) RemovePlayer(name string) {
	for i, player := range l.Players {
		if player.Name == name {
			l.Players = slices.Delete(l.Players, i, i+1)
			break
		}
	}
}

func (l *Lobby) Ready() bool {
	return len(l.Players) > 2
}

func (l *Lobby) ToBytes() ([]byte, error) {
	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func FromBytes(b []byte) (*Lobby, error) {
	lobby := &Lobby{}
	if err := json.Unmarshal(b, &lobby); err != nil {
		return nil, err
	}

	return lobby, nil
}
