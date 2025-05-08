package player

import (
	"encoding/json"
)

type Player struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
	Host  bool   `json:"host"`
}

func (p *Player) ToBytes() []byte {
	b, _ := json.Marshal(p)
	return b
}
