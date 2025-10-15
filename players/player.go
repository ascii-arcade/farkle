package players

import (
	"context"

	"github.com/ascii-arcade/farkle/database"
	"github.com/charmbracelet/ssh"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Player struct {
	SshPubKey string `bson:"ssh_pub_key"`
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

func (p *Player) SetPubKey(pKey string) {
	p.SshPubKey = pKey
}

func (p *Player) Save() error {
	if p.SshPubKey == "" {
		return nil
	}

	opts := options.Replace().SetUpsert(true)
	_, err := database.GetDB().Collection(database.CollectionPlayers).ReplaceOne(p.ctx, bson.D{{Key: "ssh_pub_key", Value: p.SshPubKey}}, p, opts)
	return err
}

func (p *Player) SetLanguage(lang string) {
	p.LanguagePreference = lang
	_ = p.Save()
}
