package game

import (
	"encoding/json"
)

type GameDetails struct {
	LobbyCode string
	PlayerId  string
	DieHeld   int
}

func (gd *GameDetails) ToJSON() string {
	data, err := json.Marshal(gd)
	if err != nil {
		return ""
	}
	return string(data)
}
