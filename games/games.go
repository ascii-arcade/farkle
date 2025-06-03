package games

import (
	"math/rand/v2"

	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/utils"
	"github.com/charmbracelet/lipgloss"
)

var games = make(map[string]*Game)

func New(style lipgloss.Style) *Game {
	colors := []string{
		"#3B82F6", // Blue
		"#10B981", // Green
		"#FACC15", // Yellow
		"#8B5CF6", // Purple
		"#06B6D4", // Cyan
		"#F97316", // Orange
	}

	rand.Shuffle(len(colors), func(i, j int) {
		colors[i], colors[j] = colors[j], colors[i]
	})

	game := &Game{
		turn:      0,
		DicePool:  dice.NewDicePool(6),
		DiceHeld:  dice.NewDicePool(0),
		FirstRoll: true,
		Code:      utils.GenerateCode(),
		style:     style,
		colors:    colors,
	}
	games[game.Code] = game
	return game
}

func Exists(code string) bool {
	_, ok := games[code]
	return ok
}

func Get(code string) (*Game, bool) {
	game, ok := games[code]
	return game, ok
}

func GetAll() []*Game {
	gamesList := make([]*Game, 0, len(games))
	for _, game := range games {
		gamesList = append(gamesList, game)
	}
	return gamesList
}

func GetOpenGame(code string) (*Game, error) {
	game, exists := games[code]
	if !exists {
		return nil, ErrGameNotFound
	}
	if game.InProgress {
		return nil, ErrGameAlreadyInProgress
	}

	return game, nil
}
