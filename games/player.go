package games

import (
	"github.com/charmbracelet/lipgloss"
)

type Player struct {
	Id             string
	Name           string
	Score          int
	Color          string
	UpdateChan     chan struct{}
	Host           bool
	PlayedLastTurn bool
}

func (p *Player) StyledPlayerName(style lipgloss.Style) string {
	return style.Foreground(lipgloss.Color(p.Color)).Render(p.Name)
}
