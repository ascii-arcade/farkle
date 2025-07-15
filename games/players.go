package games

import (
	"context"

	"github.com/ascii-arcade/farkle/language"
	"github.com/charmbracelet/ssh"
)

var players = make(map[string]*Player)

func NewPlayer(ctx context.Context, sess ssh.Session, langPref *language.LanguagePreference) *Player {
	player, exists := players[sess.User()]
	if exists {
		// player.IsHost = false
		player.UpdateChan = make(chan struct{})
		player.connected = true
		player.ctx = ctx

		goto RETURN
	}

	player = &Player{
		// Score:              0,
		// PlayedLastTurn:     false,
		UpdateChan:         make(chan struct{}),
		LanguagePreference: langPref,
		connected:          true,
		Sess:               sess,
		onDisconnect:       []func(){},
		ctx:                ctx,
	}
	players[sess.User()] = player

RETURN:
	go func() {
		<-player.ctx.Done()
		player.connected = false
		for _, fn := range player.onDisconnect {
			fn()
		}
	}()

	return player
}

func RemovePlayer(player *Player) {
	if _, exists := players[player.Sess.User()]; exists {
		close(player.UpdateChan)
		delete(players, player.Sess.User())
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
