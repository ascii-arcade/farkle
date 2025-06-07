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

	connected bool

	UpdateChan         chan struct{}
	LanguagePreference *language.LanguagePreference
	Sess               ssh.Session
	onDisconnect       []func()

	ctx context.Context
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
	if p == nil {
		return ""
	}
	return style.Foreground(p.Color).Render(p.Name)
}

func (p *Player) OnDisconnect(fn func()) {
	p.onDisconnect = append(p.onDisconnect, fn)
}
