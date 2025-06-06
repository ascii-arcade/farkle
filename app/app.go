package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"

	"github.com/ascii-arcade/farkle/games"
	"github.com/ascii-arcade/farkle/language"
	"github.com/ascii-arcade/farkle/menu"
	"github.com/ascii-arcade/farkle/messages"
)

type Model struct {
	active tea.Model
	sess   ssh.Session
}

func (m Model) Init() tea.Cmd {
	return m.active.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SwitchViewMsg:
		m.active = msg.Model
		initcmd := m.active.Init()
		return m, initcmd
	}

	var cmd tea.Cmd
	m.active, cmd = m.active.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.active.View()
}

func TeaHandler(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := sess.Pty()
	renderer := bubbletea.MakeRenderer(sess)
	style := renderer.NewStyle()

	languagePreference := language.LanguagePreference{Lang: language.DefaultLanguage}

	player := games.NewPlayer(sess.Context(), sess, &languagePreference)

	return Model{
		sess:   sess,
		active: menu.New(pty.Window.Width, pty.Window.Height, style, player),
	}, []tea.ProgramOption{tea.WithAltScreen()}
}
