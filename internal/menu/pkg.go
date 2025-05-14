package menu

import (
	"log/slog"
	"time"

	"github.com/ascii-arcade/farkle/internal/lobbies"
	"github.com/ascii-arcade/farkle/internal/message"
	"github.com/ascii-arcade/farkle/internal/player"
)

type tick time.Time

var (
	logger       *slog.Logger
	currentLobby *lobbies.Lobby
	me           *player.Player
	serverHealth bool
	messages     chan message.Message
)
