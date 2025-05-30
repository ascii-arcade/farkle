package games

import (
	"github.com/charmbracelet/lipgloss"
)

type Player struct {
	Id         string
	Name       string
	Score      int
	Color      string
	TurnOrder  int
	UpdateChan chan any
	Host       bool
}

func (p *Player) StyledPlayerName(style lipgloss.Style) string {
	return style.Foreground(lipgloss.Color(p.Color)).Render(p.Name)
}
