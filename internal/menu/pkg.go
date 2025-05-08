package menu

import (
	"time"

	"github.com/ascii-arcade/farkle/internal/lobby"
	"github.com/ascii-arcade/farkle/internal/player"
)

type tick time.Time

var wsClient *client

var lobbies = []*lobby.Lobby{}
var currentLobbyId string
var myPlayer *player.Player
var serverHealth bool

func getLobby(id string) *lobby.Lobby {
	for _, l := range lobbies {
		if l.Id == id {
			return l
		}
	}
	return nil
}

func updateLobby(lobby *lobby.Lobby) {
	for i, l := range lobbies {
		if l.Id == lobby.Id {
			lobbies[i].Players = lobby.Players
			lobbies[i].Started = lobby.Started
			return
		}
	}
	lobbies = append(lobbies, lobby)
}
