package tui

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

func stylePool() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(36).
		Height(12).
		Align(lipgloss.Center)
}

func (m model) View() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height)

	poolRollPane := stylePool().Render(m.poolRoll.render(0, 3) + "\n\n" + m.poolRoll.render(3, 6))
	poolHeldPane := stylePool().Render(m.poolHeld.render(0, 3) + "\n\n" + m.poolHeld.render(3, 6))

	centeredText := "Locked In: " + strconv.Itoa(m.lockedInScore)
	if m.error != "" {
		centeredText = lipgloss.NewStyle().Foreground(lipgloss.Color("#9E1A1A")).Render(m.error)
	}

	poolPanes := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			poolRollPane,
			poolHeldPane,
		),
		"",
		centeredText,
	)

	panes := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			poolPanes,
			"",
			m.playerScores(),
			"",
			"",
		),
		"r to roll, l to lock, n to bust, y to bank, u to undo",
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
