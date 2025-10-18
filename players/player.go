package players

import (
	"context"
	"time"

	"github.com/ascii-arcade/farkle/database"
	"github.com/ascii-arcade/farkle/utils"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Player struct {
	Id                 string            `bson:"_id,omitempty"`
	Username           string            `bson:"username"`
	Discriminator      string            `bson:"discriminator"`
	SshPubKeys         map[string]string `bson:"ssh_pub_keys"`
	LanguagePreference string            `bson:"language_preference"`
	LastConnectedAt    *time.Time        `bson:"last_connected_at,omitempty"`

	sess         ssh.Session
	updateChan   chan struct{}
	onDisconnect []func()
	connected    bool
	ctx          context.Context
}

func (p *Player) WithContext(ctx context.Context) *Player {
	p.ctx = ctx
	return p
}

func (p *Player) GetDisplayName(style lipgloss.Style) string {
	return style.Foreground(lipgloss.Color("#1dc42bff")).Render(p.Username + "#" + p.Discriminator)
}

func (p *Player) OnDisconnect(fn func()) {
	p.onDisconnect = append(p.onDisconnect, fn)
}

func (p *Player) IsConnected() bool {
	return p.connected
}

func (p *Player) AddPubKey(name, pKey string) {
	p.SshPubKeys[name] = pKey
	_ = p.Save()
}

func (p *Player) SetSession(sess ssh.Session) {
	p.sess = sess
}

func (p *Player) Save() error {
	if len(p.SshPubKeys) == 0 {
		return nil
	}

	if _, exists := GetByName(p.Username, p.Discriminator); exists {
		p.Discriminator = utils.GenerateDescriminator()
	}

	opts := options.Replace().SetUpsert(true)
	_, err := database.GetDB().Collection(database.CollectionPlayers).ReplaceOne(p.ctx, bson.M{"_id": p.Id}, p, opts)
	return err
}

func (p *Player) SetLanguage(lang string) {
	p.LanguagePreference = lang
	_ = p.Save()
}

func (p *Player) UpdateChan() chan struct{} {
	if p.updateChan == nil {
		p.updateChan = make(chan struct{}, 1)
	}
	return p.updateChan
}
