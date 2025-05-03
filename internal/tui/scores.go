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
		content := player.name + ": " + strconv.Itoa(player.score)

		if i == m.currentPlayerIndex {
			scores[i] = content
			scores[i] = styleScore().Bold(true).Foreground(lipgloss.Color(colorCurrentTurn)).Render(content)
		} else {
			scores[i] = styleScore().Render(content)
		}
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		scores...,
	)
}
