package tui

import (
	"math/rand/v2"
	"time"

	"github.com/ascii-arcade/farkle/internal/lobbies"
	"github.com/ascii-arcade/farkle/internal/player"
	tea "github.com/charmbracelet/bubbletea"
)

type log []string

type model struct {
	currentPlayerIndex int
	error              string
	isRolling          bool
	justRolled         bool
	lockedInScore      int
	log                log
	playerColors       []string
	// players            []player.Player
	poolHeld      dicePool
	poolRoll      dicePool
	rollTickCount int

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

var players []*player.Player

func (m model) Init() tea.Cmd {
	return tea.Tick(16*time.Millisecond+6*time.Microsecond, func(t time.Time) tea.Msg {
		return tick(t)
	})
}

func Run(playerNames []string) {
	players = []*player.Player{}
	for _, name := range playerNames {
		players = append(players, &player.Player{Name: name})
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
			// players:      players,
			poolHeld: newDicePool(0),
			poolRoll: newDicePool(6),
		},
		tea.WithAltScreen(),
	).Run(); err != nil {
		panic(err)
	}
}

func RunFromLobby(l *lobbies.Lobby) {
	players = l.Players
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
			// players:      players,
			poolHeld: newDicePool(0),
			poolRoll: newDicePool(6),
		},
	).Run(); err != nil {
		panic(err)
	}
}
