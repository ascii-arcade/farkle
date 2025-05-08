package tui

import (
	"math/rand/v2"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Player struct {
	Name  string
	Score int
	Host  bool
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
	players            []Player
	poolHeld           dicePool
	poolRoll           dicePool
	rollTickCount      int

	globalTicks int
	startTime   time.Time
	tps         float64
	debug       bool

	height int
	width  int
}

type tick time.Time

const (
	rollFrames   = 15
	rollInterval = 200 * time.Millisecond

	colorCurrentTurn = "#FF9E1A"
	colorError       = "#9E1A1A"
)

var players []Player

func (m model) Init() tea.Cmd {
	return tea.Tick(16*time.Millisecond+6*time.Microsecond, func(t time.Time) tea.Msg {
		return tick(t)
	})
}

func Run(playerNames []string, debug bool) {
	players = []Player{}
	for _, name := range playerNames {
		players = append(players, Player{Name: name})
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

	if _, err := tea.NewProgram(
		model{
			playerColors: colors,
			players:      players,
			poolHeld:     newDicePool(0),
			poolRoll:     newDicePool(6),
			debug:        debug,
			startTime:    time.Now(),
			tps:          0,
		},
		tea.WithAltScreen(),
	).Run(); err != nil {
		panic(err)
	}
}
