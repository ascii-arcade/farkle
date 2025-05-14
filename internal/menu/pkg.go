package menu

import (
	"log/slog"
	"time"

	"github.com/ascii-arcade/farkle/internal/lobbies"
	"github.com/ascii-arcade/farkle/internal/player"
)

type tick time.Time
type errorMsg error

var (
	logger       *slog.Logger
	currentLobby *lobbies.Lobby
	me           *player.Player
	serverHealth bool
)
