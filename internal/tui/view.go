package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func stylePool(height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(66).
		Height(height).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder())
}

func (m model) View() string {
	style := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).
		Height(m.height)

	panes := lipgloss.JoinVertical(
		lipgloss.Left,
		stylePool(5).Render(m.poolRoll.render(false)),
		stylePool(7).Render(m.poolHeld.render(true)),
		stylePool(7).Render(m.poolLocked.render(true)),
		// viewScores,
		// viewLog,
	)

	return style.Render(panes)
}
