package games

import (
	"github.com/charmbracelet/lipgloss"
)

type Player struct {
	Id             string
	Name           string
	Score          int
	Color          lipgloss.Color
	UpdateChan     chan struct{}
	Host           bool
	PlayedLastTurn bool
}

func (p *Player) StyledPlayerName(style lipgloss.Style) string {
	return style.Foreground(p.Color).Render(p.Name)
}
