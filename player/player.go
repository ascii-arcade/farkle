package player

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/xid"
)

type Player struct {
	Id        string
	Name      string
	Score     int
	Host      bool
	Active    bool
	Color     string
	TurnOrder int
}

func New(name string) Player {
	c := Player{
		Id:   xid.New().String(),
		Name: name,
	}
	return c
}

func (p *Player) StyledPlayerName() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Color))

	return style.Render(p.Name)
}
