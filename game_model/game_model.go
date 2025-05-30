package gamemodel

import (
	"time"

	"github.com/ascii-arcade/farkle/games"
	"github.com/ascii-arcade/farkle/messages"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width  int
	height int

	rollTickCount int
	error         string

	GameCode string

	style  lipgloss.Style
	player *games.Player
	game   *games.Game
}

const (
	rollFrames   = 15
	rollInterval = 200 * time.Millisecond

	colorCurrentTurn = "#FF9E1A"
	colorError       = "#9E1A1A"
)

func NewModel(style lipgloss.Style, width, height int, player *games.Player, game *games.Game) Model {
	return Model{
		player: player,
		game:   game,
		style:  style,
		width:  width,
		height: height,
	}
}

type rollMsg struct{}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		waitForRefreshSignal(m.player.UpdateChan),
		tea.WindowSize(),
	)
}

func waitForRefreshSignal(ch chan any) tea.Cmd {
	return func() tea.Msg {
		return messages.RefreshGame(<-ch)
	}
}
