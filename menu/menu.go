package menu

import (
	"time"

	"github.com/ascii-arcade/farkle/colors"
	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/language"
	"github.com/ascii-arcade/farkle/messages"
	"github.com/ascii-arcade/farkle/screen"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Width  int
	Height int

	screen             screen.Screen
	style              lipgloss.Style
	languagePreference *language.LanguagePreference

	error string
}

func New(width, height int, style lipgloss.Style, languagePreference *language.LanguagePreference) *Model {
	m := &Model{
		Width:  width,
		Height: height,

		style:              style,
		languagePreference: languagePreference,
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
		textinput.Blink,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width, m.Height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
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
	if m.Width < config.MinimumWidth {
		return m.lang().Get("error.window_too_narrow")
	}
	if m.Height < config.MinimumHeight {
		return m.lang().Get("error.window_too_short")
	}

	style := m.style.Width(m.Width).Height(m.Height)
	paneStyle := m.style.Width(m.Width).PaddingTop(1)

	panes := lipgloss.JoinVertical(
		lipgloss.Center,
		paneStyle.Align(lipgloss.Center, lipgloss.Bottom).Foreground(colors.Logo).Height(m.Height/2).Render(m.style.Align(lipgloss.Left).Render(logo)),
		// paneStyle.Align(lipgloss.Center, lipgloss.Top).Render(lipgloss.JoinHorizontal(lipgloss.Top, m.displayDice...)),
		paneStyle.Align(lipgloss.Center, lipgloss.Top).Render(m.screen.View()),
	)

	return style.Render(panes)
}

func (m *Model) lang() *language.Language {
	if m.languagePreference == nil {
		return language.DefaultLanguage
	}
	return m.languagePreference.Lang
}
