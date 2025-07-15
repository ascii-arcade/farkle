package games

import "github.com/charmbracelet/lipgloss"

type PlayerData struct {
	Name           string
	Score          int
	Color          lipgloss.Color
	PlayedLastTurn bool

	IsHost bool

	turnOrder int
}

func (pd *PlayerData) SetName(name string) *PlayerData {
	pd.Name = name
	return pd
}

func (pd *PlayerData) MakeHost() *PlayerData {
	pd.IsHost = true
	return pd
}

func (pd *PlayerData) StyledPlayerName(style lipgloss.Style) string {
	if pd == nil {
		return ""
	}
	return style.Foreground(pd.Color).Render(pd.Name)
}
