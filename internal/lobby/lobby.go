package lobby

import (
	"encoding/json"
	"slices"
	"sync"
	"time"

	"github.com/ascii-arcade/farkle/internal/player"
	"github.com/rs/xid"
)

type Lobby struct {
	Id        string
	Name      string
	Players   []*player.Player
	Started   bool
	CreatedAt time.Time

	mu sync.Mutex
}

func NewLobby(lobbyName, hostName string) *Lobby {
	players := make([]*player.Player, 0, 6)
	host := &player.Player{
		Name:  hostName,
		Score: 0,
		Host:  true,
	}
	players = append(players, host)

	return &Lobby{
		Id:        xid.New().String(),
		Name:      lobbyName,
		Players:   players,
		Started:   false,
		CreatedAt: time.Now(),
	}
}

func (l *Lobby) AddPlayer(name string) *player.Player {
	l.mu.Lock()
	defer l.mu.Unlock()
	emptyIndex := 0
	for i, player := range l.Players {
		if player == nil {
			emptyIndex = i
			break
		}
	}
	if emptyIndex == 0 && len(l.Players) == cap(l.Players) {
		return nil
	}
	newPlayer := &player.Player{
		Id:    xid.New().String(),
		Name:  name,
		Score: 0,
	}
	l.Players[emptyIndex] = newPlayer
	return newPlayer
}

func (l *Lobby) RemovePlayer(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, player := range l.Players {
		if player.Name == name {
			l.Players = slices.Delete(l.Players, i, i+1)
			break
		}
	}
}

func (l *Lobby) Ready() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
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
