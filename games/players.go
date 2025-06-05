package games

import (
	"context"

	"github.com/ascii-arcade/farkle/language"
	"github.com/ascii-arcade/farkle/utils"
	"github.com/charmbracelet/ssh"
)

var players = make(map[string]*Player)

func NewPlayer(ctx context.Context, sess ssh.Session, langPref *language.LanguagePreference) *Player {
	player, exists := players[sess.User()]
	if exists {
		player.UpdateChan = make(chan struct{})
		player.connected = true

		goto RETURN
	}

	player = &Player{
		Name:               utils.GenerateName(langPref.Lang),
		UpdateChan:         make(chan struct{}),
		LanguagePreference: langPref,
		connected:          true,
		sess:               sess,
		ctx:                ctx,
	}
	players[sess.User()] = player

RETURN:
	go func() {
		<-player.ctx.Done()
		player.connected = false
	}()

	return player
}

func RemovePlayer(player *Player) {
	if _, exists := players[player.sess.User()]; exists {
		close(player.UpdateChan)
		delete(players, player.sess.User())
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
