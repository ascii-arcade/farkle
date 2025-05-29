package gamemodel

import (
	"time"

	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/player"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Player   player.Player
	Term     string
	Width    int
	Height   int
	Renderer *lipgloss.Renderer

	poolRoll      dice.DicePool
	rolling       bool
	rollTickCount int
	error         string

	GameCode string
	UpdateCh chan any
}

const (
	rollFrames   = 15
	rollInterval = 200 * time.Millisecond

	colorCurrentTurn = "#FF9E1A"
	colorError       = "#9E1A1A"
)

func NewModel(player player.Player) Model {
	return Model{
		poolRoll: dice.NewDicePool(6),
		Player:   player,
	}
}

type rollMsg struct{}

func (m Model) Init() tea.Cmd {
	return tea.WindowSize()
}
