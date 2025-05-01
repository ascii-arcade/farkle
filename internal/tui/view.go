package tui

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

func stylePool() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(66).
		Height(5).
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
		stylePool().Render(m.poolRoll.render()),
		stylePool().Render(m.poolHeld.render()),
		strconv.Itoa(m.lockedInScore),
		// viewScores,
		// viewLog,
	)

	return style.Render(panes)
}
