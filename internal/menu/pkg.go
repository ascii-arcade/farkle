package menu

import (
	"time"

	"github.com/ascii-arcade/farkle/internal/lobby"
)

type tick time.Time

var wsClient *client

var lobbies = []*lobby.Lobby{}
