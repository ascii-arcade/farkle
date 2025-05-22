package client

import (
	"github.com/ascii-arcade/farkle/internal/client/messages"
	"github.com/ascii-arcade/farkle/internal/client/networkmanager"
	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	CurrentView    tea.Model
	NetworkManager *networkmanager.NetworkManager
}

func (m App) Init() tea.Cmd {
	return m.CurrentView.Init()
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updateModel, cmd := m.CurrentView.Update(msg)

	switch msg := msg.(type) {
	case messages.SwitchViewMsg:
		updateModel = msg.NewModel
		cmd = nil
	}

	m.CurrentView = updateModel
	return m, cmd
}

func (m App) View() string {
	return m.CurrentView.View()
}
