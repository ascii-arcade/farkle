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
	isRolling          bool
	players            []player
	poolHeld           dicePool
	poolLocked         dicePool
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

func Run() {
	tea.NewProgram(
		model{
			players:    []player{},
			poolHeld:   newDicePool(0),
			poolLocked: newDicePool(0),
			poolRoll:   newDicePool(6),
		},
		tea.WithAltScreen(),
	).Run()
}
