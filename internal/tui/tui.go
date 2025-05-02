package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type player struct {
	name  string
	score int
}

type model struct {
	currentPlayerIndex int
	error              string
	isRolling          bool
	lockedInScore      int
	players            []player
	poolHeld           dicePool
	poolRoll           dicePool
	tickCount          int

	height int
	width  int
}

const (
	rollFrames   = 15
	rollInterval = 200 * time.Millisecond
)

func (m model) Init() tea.Cmd {
	return nil
}

func Run(playerNames []string) {
	players := make([]player, len(playerNames))
	for i, name := range playerNames {
		players[i] = player{name: name}
	}

	tea.NewProgram(
		model{
			players:  players,
			poolHeld: newDicePool(0),
			poolRoll: newDicePool(6),
		},
		tea.WithAltScreen(),
	).Run()
}
