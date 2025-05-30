package splashscreen

import (
	"time"

	"github.com/ascii-arcade/farkle/menu"
	"github.com/ascii-arcade/farkle/messages"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	height int
	width  int
	style  lipgloss.Style
}

type doneMsg struct{}

func NewModel(style lipgloss.Style, width, height int) Model {
	return Model{
		height: height,
		width:  width,
		style:  style,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return doneMsg{}
		}),
		tea.WindowSize(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doneMsg:
		menuModel := menu.New(m.style, m.width, m.height)
		return m, func() tea.Msg {
			return messages.SwitchViewMsg{Model: menuModel}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m Model) View() string {
	style := m.style.
		Width(m.width).
		Height(m.height)

	logo := `++------------------------------------------------------------------------------++
++------------------------------------------------------------------------------++
||                                                                              ||
||                                                                              ||
||      _    ____   ____ ___ ___        _    ____   ____    _    ____  _____    ||
||     / \  / ___| / ___|_ _|_ _|      / \  |  _ \ / ___|  / \  |  _ \| ____|   ||
||    / _ \ \___ \| |    | | | |_____ / _ \ | |_) | |     / _ \ | | | |  _|     ||
||   / ___ \ ___) | |___ | | | |_____/ ___ \|  _ <| |___ / ___ \| |_| | |___    ||
||  /_/   \_\____/ \____|___|___|   /_/   \_\_| \_\\____/_/   \_\____/|_____|   ||
||                                                                              ||
||                                                                              ||
||                                                                              ||
++------------------------------------------------------------------------------++
++------------------------------------------------------------------------------++`

	return style.Render(
		lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			logo,
		),
	)
}
