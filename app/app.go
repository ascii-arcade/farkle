package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"

	"github.com/ascii-arcade/farkle/menu"
	"github.com/ascii-arcade/farkle/messages"
)

type rootModel struct {
	active tea.Model
	sess   ssh.Session
}

func (m rootModel) Init() tea.Cmd {
	return m.active.Init()
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m rootModel) View() string {
	return m.active.View()
}

func TeaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, active := s.Pty()
	if !active {
		return nil, nil
	}
	return rootModel{
		sess:   s,
		active: menu.New(pty.Window.Width, pty.Window.Height, bubbletea.MakeRenderer(s).NewStyle()),
	}, []tea.ProgramOption{tea.WithAltScreen()}
}
