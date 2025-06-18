package games

import (
	"context"
	"errors"

	"github.com/ascii-arcade/farkle/language"
	"github.com/charmbracelet/ssh"
)

var players = make(map[string]*Player)

var (
	ErrAlreadyConnected = errors.New("player already connected")
)

func NewPlayer(ctx context.Context, sess ssh.Session, langPref *language.LanguagePreference) (*Player, error) {
	player, exists := players[sess.User()]
	if exists {
		if player.connected {
			return nil, ErrAlreadyConnected
		}

		player.UpdateChan = make(chan struct{})
		player.connected = true
		player.ctx = ctx

		goto RETURN
	}

	player = &Player{
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

	return player, nil
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
