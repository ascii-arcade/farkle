package menu

import (
	"time"

	"github.com/ascii-arcade/farkle/internal/lobby"
	"github.com/ascii-arcade/farkle/internal/player"
)

type tick time.Time

var (
	wsClient     *client
	currentLobby *lobby.Lobby
	me           *player.Player
	serverHealth bool
	debug        bool
	serverURL    string = "localhost"
	serverPort   string = "8080"
)
