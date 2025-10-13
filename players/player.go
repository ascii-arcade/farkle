package players

import (
	"context"

	"github.com/ascii-arcade/farkle/language"
	"github.com/charmbracelet/ssh"
)

type Player struct {
	connected bool

	UpdateChan         chan struct{}
	LanguagePreference *language.LanguagePreference
	Sess               ssh.Session
	onDisconnect       []func()

	ctx context.Context
}

func (p *Player) OnDisconnect(fn func()) {
	p.onDisconnect = append(p.onDisconnect, fn)
}

func (p *Player) IsConnected() bool {
	return p.connected
}
