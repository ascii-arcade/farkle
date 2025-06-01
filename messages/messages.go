package messages

import (
	"github.com/ascii-arcade/farkle/screen"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	SwitchViewMsg   struct{ Model tea.Model }
	SwitchScreenMsg struct{ Screen screen.Screen }
	RefreshGame     any
)
