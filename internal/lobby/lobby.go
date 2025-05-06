package lobby

import "github.com/ascii-arcade/farkle/internal/tui"

type Lobby struct {
	players map[string]*tui.Player
	started bool
}

func NewLobby(playerCount int, names []string) *Lobby {
	players := make(map[string]*tui.Player, playerCount)
	for _, name := range names {
		players[name] = &tui.Player{
			Name:  name,
			Score: 0,
		}
	}

	return &Lobby{
		players: make(map[string]*tui.Player, playerCount),
		started: false,
	}
}
