package game

import (
	"encoding/json"

	"github.com/ascii-arcade/farkle/internal/message"
)

type GameDetails struct {
	LobbyCode string
	PlayerId  string
	DieHeld   int
}

func (g *Game) HandleMessage(msg message.Message) {
	details := GameDetails{}
	if err := json.Unmarshal([]byte(msg.Data.(string)), &details); err != nil {
		return
	}

	switch msg.Type {
	case message.MessageTypeRoll:
		g.RollDice()
	case message.MessageTypeHold:
		g.HoldDie(details.DieHeld)
	case message.MessageTypeUndo:
		g.Undo()
	case message.MessageTypeLock:
		g.LockDice()
	default:
	}
}

func (gd *GameDetails) ToJSON() string {
	data, err := json.Marshal(gd)
	if err != nil {
		return ""
	}
	return string(data)
}
