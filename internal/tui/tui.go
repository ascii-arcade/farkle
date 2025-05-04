package tui

import (
	"math/rand/v2"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type player struct {
	name  string
	score int
}

type log []string

type model struct {
	currentPlayerIndex int
	error              string
	isRolling          bool
	justRolled         bool
	lockedInScore      int
	log                log
	playerColors       []string
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

	colorCurrentTurn = "#FF9E1A"
	colorError       = "#9E1A1A"
)

func (m model) Init() tea.Cmd {
	return nil
}

func Run(playerNames []string) {
	players := make([]player, len(playerNames))
	for i, name := range playerNames {
		players[i] = player{name: name}
	}

	colors := []string{
		"#3B82F6", // Blue
		"#10B981", // Green
		"#FACC15", // Yellow
		"#8B5CF6", // Purple
		"#06B6D4", // Cyan
		"#F97316", // Orange
	}

	// Shuffle
	rand.Shuffle(len(colors), func(i, j int) {
		colors[i], colors[j] = colors[j], colors[i]
	})

	tea.NewProgram(
		model{
			playerColors: colors,
			players:      players,
			poolHeld:     newDicePool(0),
			poolRoll:     newDicePool(6),
		},
		tea.WithAltScreen(),
	).Run()
}
