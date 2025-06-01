package screen

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Screen interface {
	WithModel(any) Screen
	Update(tea.Msg) (any, tea.Cmd)
	View() string
}
