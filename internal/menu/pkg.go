package menu

import (
	"time"

	"github.com/ascii-arcade/farkle/internal/lobby"
	"github.com/ascii-arcade/farkle/internal/player"
)

type tick time.Time

var wsClient *client

var currentLobby *lobby.Lobby
var myPlayer *player.Player
var serverHealth bool
var debug bool
