package messages

import (
	tea "github.com/charmbracelet/bubbletea"
)

type SwitchViewMsg struct {
	NewModel tea.Model
}
