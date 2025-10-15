package players

import (
	"context"

	"github.com/ascii-arcade/farkle/database"
)

var players = make(map[string]*Player)

func NewPlayer(ctx context.Context, pk, langPref string) (*Player, error) {
	player, exists := players[pk]
	if exists {
		player.UpdateChan = make(chan struct{})
		player.connected = true
		player.ctx = ctx

		goto RETURN
	}

	player = &Player{
		SshPubKey:          pk,
		UpdateChan:         make(chan struct{}),
		LanguagePreference: langPref,
		connected:          true,
		onDisconnect:       []func(){},
		ctx:                ctx,
	}
	players[pk] = player

RETURN:
	go func() {
		<-player.ctx.Done()
		player.connected = false
		for _, fn := range player.onDisconnect {
			fn()
		}
	}()

	if err := player.Save(); err != nil {
		return nil, err
	}

	return player, nil
}

func Get(sshPubKey string) (*Player, bool) {
	var player Player
	err := database.GetDB().Collection(database.CollectionPlayers).FindOne(context.Background(), map[string]any{"ssh_pub_key": sshPubKey}).Decode(&player)
	return &player, err == nil
}

func RemovePlayer(player *Player) {
	if _, exists := players[player.SshPubKey]; exists {
		close(player.UpdateChan)
		delete(players, player.SshPubKey)
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
