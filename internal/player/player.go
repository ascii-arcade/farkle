package player

import (
	"encoding/json"

	"github.com/rs/xid"
)

type Player struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
	Host  bool   `json:"host"`
}

func New(name string) *Player {
	return &Player{
		Id:    xid.New().String(),
		Name:  name,
		Score: 0,
	}
}

func (p *Player) ToBytes() []byte {
	b, _ := json.Marshal(p)
	return b
}
