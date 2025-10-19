package games

import "github.com/charmbracelet/lipgloss"

type PlayerData struct {
	Name           string
	Score          int
	Color          lipgloss.Color
	PlayedLastTurn bool
	InGame         bool

	IsHost bool

	turnOrder int
}

func (pd *PlayerData) StyledPlayerName(style lipgloss.Style) string {
	if pd == nil {
		return ""
	}
	return style.Foreground(pd.Color).Render(pd.Name)
}
