package games

import (
	"context"

	"github.com/ascii-arcade/farkle/language"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

type Player struct {
	Name           string
	Score          int
	Color          lipgloss.Color
	PlayedLastTurn bool

	IsHost bool

	UpdateChan         chan struct{}
	LanguagePreference *language.LanguagePreference

	sess ssh.Session
	ctx  context.Context
}

func (p *Player) SetName(name string) *Player {
	p.Name = name
	return p
}

func (p *Player) MakeHost() *Player {
	p.IsHost = true
	return p
}

func (p *Player) StyledPlayerName(style lipgloss.Style) string {
	return style.Foreground(p.Color).Render(p.Name)
}
