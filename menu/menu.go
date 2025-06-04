package menu

import (
	"time"

	"github.com/ascii-arcade/farkle/colors"
	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/games"
	"github.com/ascii-arcade/farkle/language"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/screen"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const logo = `
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⡆⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⡟⠋⠛⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣧⣀⣠⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⡆⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⡆⠀⠀⠀⠀
⠀⠀⠀⠀⣿⣿⡟⠋⠛⣿⣿⣿⣿⣿⡟⠋⠛⣿⣿⡇⣿⣿⡟⠋⠛⣿⣿⣿⣿⣿⡟⠋⠛⣿⣿⡇⠀⠀⠀⠀
⠀⠀⠀⠀⣿⣿⣧⣀⣀⣾⣿⣿⣿⣿⣧⣀⣀⣾⣿⡇⣿⣿⣧⣀⣀⣾⣿⣿⣿⣿⣧⣀⣀⣾⣿⡇⠀⠀⠀⠀
⠀⠀⠀⠀⣿⣿⡟⠋⠛⣿⣿⣿⣿⣿⡟⠋⠛⣿⣿⡇⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀
⠀⠀⠀⠀⣿⣿⣧⣀⣀⣾⣿⣿⣿⣿⣧⣀⣀⣾⣿⡇⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀
⠀⠀⠀⠀⣿⣿⡟⠋⠛⣿⣿⣿⣿⣿⡟⠋⠛⣿⣿⡇⣿⣿⡟⠋⠛⣿⣿⣿⣿⣿⡟⠋⠛⣿⣿⡇⠀⠀⠀⠀
⠀⠀⠀⠀⣿⣿⣧⣀⣀⣾⣿⣿⣿⣿⣧⣀⣀⣾⣿⡇⣿⣿⣧⣀⣀⣾⣿⣿⣿⣿⣧⣀⣀⣾⣿⡇⠀⠀⠀⠀
⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀`

type Model struct {
	width  int
	height int

	screen screen.Screen
	style  lipgloss.Style

	player *games.Player

	error string
}

func New(width, height int, style lipgloss.Style, player *games.Player) *Model {
	m := &Model{
		width:  width,
		height: height,

		style: style,

		player: player,
	}
	m.screen = m.newSplashScreen()
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return messages.SplashScreenDoneMsg{}
		}),
		tea.WindowSize(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			games.RemovePlayer(m.player)
			return m, tea.Quit
		}

	case messages.SwitchScreenMsg:
		m.screen = msg.Screen.WithModel(&m)
		return m, nil
	}

	screenModel, cmd := m.screen.Update(msg)
	return screenModel.(*Model), cmd
}

func (m Model) View() string {
	if m.width < config.MinimumWidth {
		return m.lang().Get("error", "window_too_narrow")
	}
	if m.height < config.MinimumHeight {
		return m.lang().Get("error", "window_too_short")
	}

	style := m.style.Width(m.width).Height(m.height)
	paneStyle := m.style.Width(m.width).PaddingTop(1)

	panes := lipgloss.JoinVertical(
		lipgloss.Center,
		paneStyle.Align(lipgloss.Center, lipgloss.Bottom).Foreground(colors.Logo).Height(m.height/2).Render(m.style.Align(lipgloss.Left).Render(logo)),
		paneStyle.Align(lipgloss.Center, lipgloss.Top).Render(m.screen.View()),
	)

	return style.Render(panes)
}

func (m *Model) lang() *language.Language {
	return m.player.LanguagePreference.Lang
}
