package splash_screen

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	height int
	width  int
}

type doneMsg struct{}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return doneMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doneMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m model) View() string {
	style := lipgloss.NewStyle().
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

func Run() {
	tea.NewProgram(
		model{},
		tea.WithAltScreen(),
	).Run()
}
