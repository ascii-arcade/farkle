package lobby

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/ascii-arcade/farkle/internal/player"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/rs/xid"
)

type Lobby struct {
	Id        string           `json:"id"`
	Name      string           `json:"name"`
	Players   []*player.Player `json:"players"`
	Started   bool             `json:"started"`
	CreatedAt time.Time        `json:"created_at"`
	Code      string           `json:"code"`

	// mu sync.Mutex `json:"-"`
}

func NewLobby(host *player.Player) *Lobby {
	players := make([]*player.Player, 6)
	host.Host = true
	players[0] = host

	return &Lobby{
		Id:        xid.New().String(),
		Name:      petname.Generate(2, "-"),
		Code:      newCode(),
		Players:   players,
		Started:   false,
		CreatedAt: time.Now(),
	}
}

func newCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b[:3]) + "-" + string(b[3:6])
}

// func FromMap(m map[string]any) (*Lobby, error) {
// 	b, err := json.Marshal(m)
// 	if err != nil {
// 		return nil, err
// 	}

// 	lobby := &Lobby{}
// 	if err := json.Unmarshal(b, &lobby); err != nil {
// 		return nil, err
// 	}

// 	return lobby, nil
// }

func (l *Lobby) AddPlayer(p *player.Player) bool {
	// l.mu.Lock()
	// defer l.mu.Unlock()
	emptyIndex := 0
	for i, player := range l.Players {
		if player == nil {
			emptyIndex = i
			break
		}

		if player.Id == p.Id {
			return true
		}
	}
	if emptyIndex == 0 && l.IsFull() {
		return false
	}
	l.Players[emptyIndex] = p

	return true
}

func (l *Lobby) RemovePlayer(p *player.Player) {
	// l.mu.Lock()
	// defer l.mu.Unlock()
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
	// l.mu.Lock()
	// defer l.mu.Unlock()
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
