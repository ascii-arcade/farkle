package client

import (
	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	activeModel tea.Model
}

type switchViewMsg struct {
	newModel tea.Model
}

func New() App {
	return App{
		activeModel: NewHomeModel(),
	}
}

func (m App) Init() tea.Cmd {
	return m.activeModel.Init()
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updateModel, cmd := m.activeModel.Update(msg)

	switch msg := msg.(type) {
	case switchViewMsg:
		updateModel = msg.newModel
		cmd = nil
	}

	m.activeModel = updateModel
	return m, cmd
}

func (m App) View() string {
	return m.activeModel.View()
}
