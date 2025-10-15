package players

import (
	"bytes"
	"context"

	"github.com/ascii-arcade/farkle/database"
	"github.com/charmbracelet/ssh"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	gossh "golang.org/x/crypto/ssh"
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

func (p *Player) SetSession(sess ssh.Session) {
	p.Sess = sess
}

func (p *Player) Save() error {
	if p.SshPubKey == "" {
		return nil
	}

	pk, _, _, _, err := gossh.ParseAuthorizedKey([]byte(p.SshPubKey))
	if err != nil {
		return err
	}
	decodedKey := string(bytes.TrimSuffix(gossh.MarshalAuthorizedKey(pk), []byte{'\n'}))

	opts := options.Replace().SetUpsert(true)
	_, err = database.GetDB().Collection(database.CollectionPlayers).ReplaceOne(p.ctx, bson.D{{Key: "ssh_pub_key", Value: decodedKey}}, p, opts)
	return err
}

func (p *Player) SetLanguage(lang string) {
	p.LanguagePreference = lang
	_ = p.Save()
}
