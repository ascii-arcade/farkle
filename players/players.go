package players

import (
	"context"

	"github.com/ascii-arcade/farkle/database"
	"github.com/ascii-arcade/farkle/language"
	"github.com/ascii-arcade/farkle/utils"
)

var players = make(map[string]*Player)

func NewPlayer(ctx context.Context, pk, langPref string) (*Player, error) {
	player, exists := Get(pk)
	if !exists {
		player = &Player{
			Username:           utils.GenerateName(language.Languages[langPref]),
			Discriminator:      utils.GenerateDescriminator(),
			SshPubKeys:         []string{pk},
			UpdateChan:         make(chan struct{}),
			LanguagePreference: langPref,
			connected:          true,
			onDisconnect:       []func(){},
			ctx:                ctx,
		}
	}

	player.UpdateChan = make(chan struct{})
	player.connected = true
	player.ctx = ctx
	players[player.GetDisplayName()] = player

	go func() {
		<-player.ctx.Done()
		player.connected = false
		for _, fn := range player.onDisconnect {
			fn()
		}
	}()

	return player, player.Save()
}

func Get(sshPubKey string) (*Player, bool) {
	var player Player
	err := database.GetDB().Collection(database.CollectionPlayers).FindOne(context.Background(), map[string]any{"ssh_pub_key": sshPubKey}).Decode(&player)
	return &player, err == nil
}

func RemovePlayer(player *Player) {
	if _, exists := players[player.GetDisplayName()]; exists {
		close(player.UpdateChan)
		delete(players, player.GetDisplayName())
	}
}

func GetPlayerCount() int {
	return len(players)
}

func GetConnectedPlayerCount() int {
	count := 0
	for _, player := range players {
		if player.connected {
			count++
		}
	}
	return count
}
