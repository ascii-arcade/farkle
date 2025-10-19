package games

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/ascii-arcade/farkle/database"
	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/players"
	"github.com/ascii-arcade/farkle/utils"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

var games = make(map[string]*Game)

func New(style lipgloss.Style) (*Game, error) {
	colors := []lipgloss.Color{
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
		Id:        uuid.NewString(),
		Turn:      0,
		dicePool:  dice.NewDicePool(6),
		diceHeld:  dice.NewDicePool(0),
		firstRoll: true,
		Code:      utils.GenerateCode(),
		style:     style,
		colors:    colors,
		players:   make(map[*players.Player]*PlayerData),
		CreatedAt: utils.ToPointer(time.Now()),
	}
	game.Restart()
	games[game.Code] = game

	return game, game.Save()
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
	games := make([]*Game, 0, len(games))
	cursor, err := database.GetDB().Collection(database.CollectionGames).Find(context.TODO(), bson.D{})
	if err != nil {
		return games
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var game Game
		if err := cursor.Decode(&game); err == nil {
			games = append(games, &game)
		}
	}

	return games
}

func GetOpenGame(code string) (*Game, error) {
	game, exists := games[code]
	if !exists {
		return nil, ErrGameNotFound
	}
	if game.InProgress {
		return game, ErrGameAlreadyInProgress
	}

	return game, nil
}
