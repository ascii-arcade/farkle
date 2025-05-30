package messages

import (
	tea "github.com/charmbracelet/bubbletea"
)

type (
	SwitchViewMsg struct{ Model tea.Model }
	RefreshGame   any
)
