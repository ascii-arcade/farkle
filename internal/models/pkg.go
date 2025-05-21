package client

import (
	"log/slog"
	"time"

	"github.com/ascii-arcade/farkle/internal/lobbies"
	"github.com/ascii-arcade/farkle/internal/message"
	"github.com/ascii-arcade/farkle/internal/player"
)

var (
	logger          *slog.Logger
	currentLobby    *lobbies.Lobby
	me              *player.Player
	messages        chan message.Message
	gameMessages    chan message.Message
	lastHealthCheck time.Time
)
