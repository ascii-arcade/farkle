package players

import (
	"context"

	"github.com/ascii-arcade/farkle/database"
	"github.com/charmbracelet/ssh"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Player struct {
	id        string `bson:"_id"`
	connected bool   `bson:"-"`

	UpdateChan         chan struct{} `bson:"-"`
	LanguagePreference string        `bson:"language_preference"`
	Sess               ssh.Session   `bson:"-"`
	onDisconnect       []func()

	ctx context.Context
}

func (p *Player) OnDisconnect(fn func()) {
	p.onDisconnect = append(p.onDisconnect, fn)
}

func (p *Player) IsConnected() bool {
	return p.connected
}

func (p *Player) Save() error {
	opts := options.FindOneAndReplace().SetUpsert(true)
	database.GetDB().Collection(database.CollectionPlayers).FindOneAndReplace(p.ctx, map[string]any{
		"_id": p.id,
	}, p, opts)
	return nil
}

func (p *Player) SetLanguage(lang string) {
	p.LanguagePreference = lang
	_ = p.Save()
}
