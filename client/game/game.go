package game

import (
	"time"

	"github.com/ascii-arcade/farkle/client/networkmanager"
	"github.com/ascii-arcade/farkle/dice"
	"github.com/ascii-arcade/farkle/game"
	"github.com/ascii-arcade/farkle/player"
	tea "github.com/charmbracelet/bubbletea"
)

type gameModel struct {
	width         int
	height        int
	poolRoll      dice.DicePool
	rolling       bool
	error         string
	rollTickCount int

	game   *game.Game
	player *player.Player
	nm     *networkmanager.NetworkManager
}

const (
	rollFrames   = 15
	rollInterval = 200 * time.Millisecond

	colorCurrentTurn = "#FF9E1A"
	colorError       = "#9E1A1A"
)

func NewModel(networkManager *networkmanager.NetworkManager, game *game.Game, player *player.Player) gameModel {
	return gameModel{
		poolRoll: dice.NewDicePool(6),
		game:     game,
		player:   player,
		nm:       networkManager,
	}
}

type rollMsg struct{}

func (m gameModel) Init() tea.Cmd {
	return tea.WindowSize()
}
