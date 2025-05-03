package tui

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

func styleScore() lipgloss.Style {
	return lipgloss.NewStyle().
		PaddingRight(2)
}

func (m *model) playerScores() string {
	scores := make([]string, len(m.players))

	for i, player := range m.players {
		content := m.styledPlayerName(i) + ": " + strconv.Itoa(player.score)
		isCurrentPlayer := m.currentPlayerIndex == i

		scores[i] = styleScore().
			Bold(isCurrentPlayer).
			Italic(isCurrentPlayer).
			Render(content)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		scores...,
	)
}
