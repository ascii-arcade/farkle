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

	return style.Render(players[i].Name)
}

func (m model) View() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height)

	poolRollPane := stylePool().Render(m.poolRoll.render(0, 3) + "\n" + m.poolRoll.render(3, 6))
	poolHeldPane := stylePool().Render(m.poolHeld.render(0, 3) + "\n" + m.poolHeld.render(3, 6))

	centeredText := ""
	if m.error != "" {
		centeredText = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorError)).Render(m.error)
	}

	debugPane := ""

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
			debugPane,
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
