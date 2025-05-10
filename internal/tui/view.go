package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func stylePool() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(36).
		Height(10).
		Align(lipgloss.Center)
}

func (m *model) styledPlayerName(i int) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(m.playerColors[i]))

	return style.Render(m.players[i].name)
}

func (m model) View() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height)

	poolRollPane := stylePool().Render(m.poolRoll.render())
	poolHeldPane := stylePool().Render(m.poolHeld.render())

	centeredText := ""
	if m.error != "" {
		centeredText = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorError)).Render(m.error)
	}

	poolPanes := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			poolRollPane,
			poolHeldPane,
		),
		centeredText,
	)

	panes := lipgloss.JoinVertical(
		lipgloss.Center,
		"r to roll, l to lock, n to bust, y to bank, u to undo",
		lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			poolPanes,
			m.playerScores(),
			"",
			m.logPane(),
		),
	)

	return style.Render(
		lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			panes,
		),
	)
}
