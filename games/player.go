package games

import (
	"context"

	"github.com/ascii-arcade/farkle/language"
	"github.com/ascii-arcade/farkle/utils"
	"github.com/charmbracelet/lipgloss"
)

type Player struct {
	Name           string
	Score          int
	Color          lipgloss.Color
	PlayedLastTurn bool

	IsHost bool

	UpdateChan         chan struct{}
	LanguagePreference *language.LanguagePreference

	ctx context.Context
}

func NewPlayer(ctx context.Context, langPref *language.LanguagePreference) *Player {
	return &Player{
		Name:               utils.GenerateName(langPref.Lang),
		UpdateChan:         make(chan struct{}),
		LanguagePreference: langPref,
		ctx:                ctx,
	}
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
