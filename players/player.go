package players

import (
	"context"

	"github.com/ascii-arcade/farkle/database"
	"github.com/charmbracelet/ssh"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Player struct {
	Username           string   `bson:"username"`
	Discriminator      string   `bson:"discriminator"`
	SshPubKeys         []string `bson:"ssh_pub_keys"`
	LanguagePreference string   `bson:"language_preference"`

	Sess         ssh.Session   `bson:"-"`
	UpdateChan   chan struct{} `bson:"-"`
	onDisconnect []func()
	connected    bool `bson:"-"`
	ctx          context.Context
}

func (p *Player) GetDisplayName() string {
	return p.Username + "#" + p.Discriminator
}

func (p *Player) OnDisconnect(fn func()) {
	p.onDisconnect = append(p.onDisconnect, fn)
}

func (p *Player) IsConnected() bool {
	return p.connected
}

func (p *Player) AddPubKey(pKey string) {
	p.SshPubKeys = append(p.SshPubKeys, pKey)
}

func (p *Player) SetSession(sess ssh.Session) {
	p.Sess = sess
}

func (p *Player) Save() error {
	if len(p.SshPubKeys) == 0 {
		return nil
	}

	opts := options.Replace().SetUpsert(true)
	_, err := database.GetDB().Collection(database.CollectionPlayers).ReplaceOne(p.ctx, bson.M{"username": p.Username, "discriminator": p.Discriminator}, p, opts)
	return err
}

func (p *Player) SetLanguage(lang string) {
	p.LanguagePreference = lang
	_ = p.Save()
}
