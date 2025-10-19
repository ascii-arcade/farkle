package games

import (
	"log/slog"
	"time"
)

func StartCleanup(stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}

		slog.Debug("checking for abandoned games")
		abandonedGames := 0
		for _, game := range games {
			game.withLock(func() {
				if game.GetPlayerCount(false) == 0 && game.EndedAt != nil {
					delete(games, game.Code)
					abandonedGames++
					return
				}
			})
		}

		if abandonedGames > 0 {
			slog.Debug("found abandoned games", "count", abandonedGames)
		}
		time.Sleep(5 * time.Second)
	}
}
