package lobby

import (
	"encoding/json"

	"github.com/ascii-arcade/farkle/internal/tui"
)

type Lobby struct {
	Name    string
	Players map[string]*tui.Player
	Started bool
}

func NewLobby(lobbyName string, hostName string) *Lobby {
	players := make(map[string]*tui.Player)
	players[hostName] = &tui.Player{
		Name:  hostName,
		Score: 0,
		Host:  true,
	}

	return &Lobby{
		Name:    lobbyName,
		Players: players,
	}
}

func (l *Lobby) AddPlayer(name string) {}

func (l *Lobby) RemovePlayer(name string) {
	delete(l.Players, name)
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
