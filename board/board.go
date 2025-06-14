package board

import (
	"time"

	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/games"
	"github.com/ascii-arcade/farkle/keys"
	"github.com/ascii-arcade/farkle/language"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/screen"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width  int
	height int

	error string

	style  lipgloss.Style
	player *games.Player
	game   *games.Game
	screen screen.Screen
}

const (
	rollFrames   = 15
	rollInterval = 200 * time.Millisecond
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
	case tea.KeyMsg:
		if keys.ExitApplication.TriggeredBy(msg.String()) {
			m.game.RemovePlayer(m.player)
			return m, tea.Quit
		}

	case messages.SwitchScreenMsg:
		m.screen = msg.Screen.WithModel(&m)
		return m, nil

	case messages.RefreshGame:
		return m, waitForRefreshSignal(m.player.UpdateChan)
	}

	activeScreenModel, cmd := m.activeScreen().Update(msg)
	return activeScreenModel.(*Model), cmd
}

func (m Model) View() string {
	if m.width < config.MinimumWidth {
		return m.lang().Get("error", "window_too_narrow")
	}
	if m.height < config.MinimumHeight {
		return m.lang().Get("error", "window_too_short")
	}
	return m.activeScreen().View()
}

func (m *Model) activeScreen() screen.Screen {
	return m.screen.WithModel(m)
}

func waitForRefreshSignal(ch chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return messages.RefreshGame(<-ch)
	}
}

func (m *Model) lang() *language.Language {
	return m.player.LanguagePreference.Lang
}
