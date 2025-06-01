package board

import (
	"time"

	"github.com/ascii-arcade/farkle/games"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/screen"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width  int
	height int

	rollTickCount int
	error         string

	style  lipgloss.Style
	player *games.Player
	game   *games.Game
	screen screen.Screen

	rolling bool
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
		screen: &tableScreen{},
	}
}

type rollMsg struct{}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		waitForRefreshSignal(m.player.UpdateChan),
		tea.WindowSize(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height, m.width = msg.Height, msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.game.RemovePlayer(m.player)
			return m, tea.Quit
		case "?":
			m.screen = &helpScreen{model: &m}
		case "q":
			m.screen = &tableScreen{model: &m}
		default:
			activeScreenModel, cmd := m.activeScreen().Update(msg)
			return activeScreenModel.(*Model), cmd
		}

	case rollMsg:
		if m.rollTickCount < rollFrames {
			m.rollTickCount++
			m.game.DicePool.Roll()
			m.game.Refresh()
			return m, tea.Tick(rollInterval, func(time.Time) tea.Msg {
				return rollMsg{}
			})
		}
		m.game.RollDice()
		m.rolling = false
		return m, nil

	case messages.RefreshGame:
		return m, waitForRefreshSignal(m.player.UpdateChan)
	}

	return m, nil
}

func (m Model) View() string {
	return m.activeScreen().View()
}

func (m *Model) activeScreen() screen.Screen {
	return m.screen.WithModel(m)
}

func waitForRefreshSignal(ch chan any) tea.Cmd {
	return func() tea.Msg {
		return messages.RefreshGame(<-ch)
	}
}
