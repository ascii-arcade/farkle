package menu

import (
	"log/slog"
	"time"

	"github.com/ascii-arcade/farkle/internal/lobby"
	"github.com/ascii-arcade/farkle/internal/player"
)

type tick time.Time
type errorMsg error

var (
	logger       *slog.Logger
	wsClient     *client
	currentLobby *lobby.Lobby
	me           *player.Player
	serverHealth bool
	debug        bool
	serverURL    string = "localhost"
	serverPort   string = "8080"
)
