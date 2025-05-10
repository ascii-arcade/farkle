package lobby

import (
	"encoding/json"
	"math/rand"
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
	Code      string

	mu sync.Mutex
}

func NewLobby(host *player.Player) *Lobby {
	players := make([]*player.Player, 6)
	players[0] = host

	return &Lobby{
		Id:        xid.New().String(),
		Name:      host.Name + "'s game",
		Code:      newCode(),
		Players:   players,
		Started:   false,
		CreatedAt: time.Now(),
	}
}

func newCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b[0:2]) + "-" + string(b[3:5])
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
	playerCount := 0
	for _, player := range l.Players {
		if player != nil {
			playerCount++
			if playerCount > 2 {
				return true
			}
		}
	}

	return false
}

func (l *Lobby) ToBytes() []byte {
	b, _ := json.Marshal(l)
	return b
}

func FromBytes(b []byte) (*Lobby, error) {
	lobby := &Lobby{}
	if err := json.Unmarshal(b, &lobby); err != nil {
		return nil, err
	}

	return lobby, nil
}
